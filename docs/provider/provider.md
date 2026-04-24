# Data Provider

**Package:** `pkg/provider/`, `pkg/provider/zerodha/`, `pkg/provider/zerodha/cache/`

The provider is the only place in the engine that talks to the outside world.
Everything else is pure in-memory computation.

---

## Interface

```go
// pkg/provider/provider.go
type DataProvider interface {
    FetchCandles(
        ctx context.Context,
        instrument string,     // e.g. "NSE:RELIANCE"
        tf model.Timeframe,
        from, to time.Time,
    ) ([]model.Candle, error)
}
```

The engine calls this once per `Run`. The full series is loaded into memory
before the event loop starts — there is no streaming.

---

## Zerodha Implementation Flow

```
FetchCandles("NSE:RELIANCE", daily, 2018-01-01, 2025-01-01)
        │
        ▼
┌───────────────────────────────────────────────────────┐
│  Cache lookup                                         │
│  .cache/zerodha/nse_reliance/daily_2018-01-01_2025-01-01.json │
│                                                       │
│  Cache HIT  → deserialize JSON → []Candle → done     │
│  Cache MISS → continue to API fetch                   │
└───────────────────────────┬───────────────────────────┘
                            │ MISS
                            ▼
┌───────────────────────────────────────────────────────┐
│  Instrument token lookup                              │
│  "NSE:RELIANCE" → instrument_token (e.g. 738561)     │
│                                                       │
│  instruments.csv is downloaded once and cached in     │
│  memory for the process lifetime. Contains all NSE    │
│  instruments with their integer token.                │
└───────────────────────────┬───────────────────────────┘
                            │ token
                            ▼
┌───────────────────────────────────────────────────────┐
│  Chunked API fetch                                    │
│                                                       │
│  Kite Connect allows max 2000 candles per request.    │
│  The provider automatically chunks long date ranges   │
│  into multiple sequential API calls and concatenates  │
│  the results.                                         │
│                                                       │
│  GET /instruments/historical/{token}/{interval}       │
│      ?from=2018-01-01&to=2019-12-31                   │
│  GET /instruments/historical/{token}/{interval}       │
│      ?from=2020-01-01&to=2021-12-31                   │
│  ...                                                  │
└───────────────────────────┬───────────────────────────┘
                            │ raw JSON candles
                            ▼
┌───────────────────────────────────────────────────────┐
│  Candle validation                                    │
│  model.NewCandle() validates each candle:             │
│  • High >= max(Open, Close)                           │
│  • Low  <= min(Open, Close)                           │
│  • Low  <= High                                       │
│  • Prices > 0                                         │
│  Bad candles return an error — the fetch fails fast.  │
└───────────────────────────┬───────────────────────────┘
                            │ []model.Candle
                            ▼
┌───────────────────────────────────────────────────────┐
│  Cache write                                          │
│  Serialized to JSON at the cache path.                │
│  Subsequent runs with the same instrument/tf/range    │
│  skip the API entirely.                               │
└───────────────────────────────────────────────────────┘
```

---

## Authentication Flow

Zerodha Kite Connect uses OAuth. The provider handles this automatically.

```
BuildProvider(ctx)
      │
      ▼
Load KITE_API_KEY + KITE_API_SECRET from environment (or .env file)
      │
      ▼
Check token file: ~/.config/backtest/token.json
  (override with BACKTEST_TOKEN_PATH env var)
      │
      ├── Token exists and not expired → use it
      │
      └── Token missing or expired
                │
                ▼
          Generate Kite login URL
          Print to terminal: "Open this URL in your browser: ..."
          User logs in, browser redirects to localhost with ?request_token=...
          Provider exchanges request_token for access_token
          Saves access_token to token.json with expiry
```

The access token is valid for one trading day. Once saved, it is reused
automatically until expiry.

---

## Cache

**Location:** `.cache/zerodha/` (override with `BACKTEST_CACHE_DIR`)

**Key format:** `{instrument_slug}/{timeframe}_{from}_{to}.json`

Example:
```
.cache/zerodha/
  nse_reliance/
    daily_2018-01-01_2025-01-01.json
  nse_nifty_50/
    daily_2015-01-01_2022-12-31.json
    daily_2024-01-01_2024-12-31.json
```

Cache is keyed on the exact from/to range. Requesting a slightly different
range (e.g. 2018-01-02 instead of 2018-01-01) will miss the cache and re-fetch.

**Note on corporate actions:** Zerodha returns adjusted daily candles —
prices are retroactively adjusted for splits, bonuses, and dividends.
This means the cache for a given range may contain slightly different
values after a corporate action occurs on the instrument. If you need
point-in-time correctness for pre-corporate-action data, re-fetch after the event.

---

## Timeframe Support

| Timeframe | Kite interval | Notes |
|---|---|---|
| `1min` | `minute` | High data volume; chunking is aggressive |
| `5min` | `5minute` | — |
| `15min` | `15minute` | — |
| `daily` | `day` | Default; Kite returns adjusted prices |
| `weekly` | — | Not supported by Kite Connect; returns error |

---

## Environment Variables

| Variable | Default | Purpose |
|---|---|---|
| `KITE_API_KEY` | required | Zerodha API key |
| `KITE_API_SECRET` | required | Zerodha API secret |
| `BACKTEST_TOKEN_PATH` | `~/.config/backtest/token.json` | Token storage location |
| `BACKTEST_CACHE_DIR` | `.cache/zerodha` | Candle cache root directory |
