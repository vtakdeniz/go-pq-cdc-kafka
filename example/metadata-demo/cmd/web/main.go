package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dsn := env("DATABASE_URL", "postgres://cdc_user:cdc_pass@127.0.0.1:5432/cdc_db?sslmode=disable")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("open db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		slog.Error("ping db", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", serveIndex)
	mux.HandleFunc("POST /api/insert", handleInsert(db))
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	addr := env("HTTP_ADDR", ":3000")
	slog.Info("web ui listening", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("listen", "error", err)
		os.Exit(1)
	}
}

func handleInsert(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		defer tx.Rollback() //nolint:errcheck

		stamp := time.Now().UTC().Format("15:04:05.000")
		var parentID int
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO parents (name) VALUES ($1) RETURNING id`,
			"parent-"+stamp,
		).Scan(&parentID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		var childID int
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO children (parent_id, name) VALUES ($1, $2) RETURNING id`,
			parentID, "child-"+stamp,
		).Scan(&childID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		var xid uint32
		if err := tx.QueryRowContext(ctx, `SELECT txid_current()::bigint % 4294967296`).Scan(&xid); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		if err := tx.Commit(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"parentId":      parentID,
			"childId":       childID,
			"transactionId": xid,
			"hint":          "Open Kafka UI → topic cdc.demo. Both events should share this transactionId; lsn should increase.",
		})
	}
}

func serveIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>CDC metadata demo</title>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
  <link href="https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,500;9..144,700&family=IBM+Plex+Sans:wght@400;500;600&display=swap" rel="stylesheet" />
  <style>
    :root {
      --ink: #1c2416;
      --muted: #5c6b52;
      --paper: #f3efe4;
      --panel: rgba(255, 252, 245, 0.82);
      --line: #d5cdb8;
      --accent: #0f6b5c;
      --accent-ink: #f7fff9;
      --shadow: 0 24px 60px rgba(28, 36, 22, 0.12);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: "IBM Plex Sans", sans-serif;
      color: var(--ink);
      background:
        radial-gradient(1200px 600px at 10% -10%, #dfead6 0%, transparent 55%),
        radial-gradient(900px 500px at 100% 0%, #e7d7c2 0%, transparent 50%),
        linear-gradient(160deg, #f7f3ea 0%, #ebe4d4 100%);
    }
    main {
      width: min(720px, calc(100% - 2rem));
      margin: 0 auto;
      padding: 4.5rem 0 3rem;
    }
    h1 {
      font-family: Fraunces, Georgia, serif;
      font-weight: 700;
      font-size: clamp(2.4rem, 5vw, 3.4rem);
      line-height: 1.05;
      margin: 0 0 0.75rem;
      letter-spacing: -0.02em;
    }
    .lead {
      color: var(--muted);
      font-size: 1.05rem;
      line-height: 1.55;
      max-width: 38rem;
      margin: 0 0 2rem;
    }
    .panel {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 18px;
      padding: 1.5rem;
      box-shadow: var(--shadow);
      backdrop-filter: blur(8px);
    }
    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 0.75rem;
      align-items: center;
      margin-bottom: 1.25rem;
    }
    button, a.btn {
      appearance: none;
      border: 0;
      border-radius: 999px;
      padding: 0.85rem 1.25rem;
      font: inherit;
      font-weight: 600;
      cursor: pointer;
      text-decoration: none;
      transition: transform 160ms ease, background 160ms ease, opacity 160ms ease;
    }
    button:hover, a.btn:hover { transform: translateY(-1px); }
    button:active, a.btn:active { transform: translateY(0); }
    button:disabled { opacity: 0.55; cursor: wait; transform: none; }
    button.primary {
      background: var(--accent);
      color: var(--accent-ink);
    }
    a.btn.secondary {
      background: transparent;
      color: var(--ink);
      border: 1px solid var(--line);
    }
    pre {
      margin: 0;
      padding: 1rem 1.1rem;
      border-radius: 12px;
      background: #1c2416;
      color: #e8f0df;
      font-size: 0.92rem;
      line-height: 1.45;
      overflow: auto;
      min-height: 8rem;
      white-space: pre-wrap;
    }
    .hint {
      margin-top: 1rem;
      color: var(--muted);
      font-size: 0.95rem;
      line-height: 1.5;
    }
    code {
      font-family: "IBM Plex Sans", ui-monospace, monospace;
      background: rgba(15, 107, 92, 0.08);
      padding: 0.1rem 0.35rem;
      border-radius: 6px;
    }
  </style>
</head>
<body>
  <main>
    <h1>CDC metadata demo</h1>
    <p class="lead">
      Insert a parent and child in one PostgreSQL transaction. The CDC connector
      publishes both rows to Kafka with <code>txid</code> and <code>lsn</code> headers.
    </p>
    <section class="panel">
      <div class="actions">
        <button class="primary" id="insertBtn" type="button">Insert FK pair</button>
        <a class="btn secondary" href="http://localhost:8085" target="_blank" rel="noreferrer">Open Kafka UI</a>
      </div>
      <pre id="out">Ready. Press the button, then open topic <strong>cdc.demo</strong> in Kafka UI.</pre>
      <p class="hint">
        Expect two messages with the same <code>transactionId</code>, increasing <code>lsn</code>,
        tables <code>public.parents</code> then <code>public.children</code>.
      </p>
    </section>
  </main>
  <script>
    const out = document.getElementById("out");
    const btn = document.getElementById("insertBtn");
    btn.addEventListener("click", async () => {
      btn.disabled = true;
      out.textContent = "Inserting in one transaction…";
      try {
        const res = await fetch("/api/insert", { method: "POST" });
        const body = await res.json();
        if (!res.ok) throw new Error(body.error || res.statusText);
        out.textContent = JSON.stringify(body, null, 2);
      } catch (err) {
        out.textContent = "Error: " + err.message;
      } finally {
        btn.disabled = false;
      }
    });
  </script>
</body>
</html>
`