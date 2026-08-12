const capabilities = [
  {
    number: "01",
    title: "Web experience",
    description:
      "短いモーション計測から結果表示まで、迷わず完結する体験を Next.js で検証。",
    tags: ["Next.js", "React", "TypeScript"],
  },
  {
    number: "02",
    title: "API gateway",
    description:
      "認証、タイムアウト、request ID、ヘルスチェックを Go の境界に集約。",
    tags: ["Go", "Gin", "REST API"],
  },
  {
    number: "03",
    title: "ML serving",
    description:
      "キーポイント列を前処理し、ONNX Runtime でスコアと3つの指標へ変換。",
    tags: ["Python", "FastAPI", "ONNX"],
  },
  {
    number: "04",
    title: "Cloud & ops",
    description:
      "Cloud Run、Terraform、CI/CD、SLOを通して、小さな構成の運用面まで試作。",
    tags: ["GCP", "Terraform", "Cloud Build"],
  },
];

const learnings = [
  {
    title: "境界を増やすコスト",
    body: "Next.js・Go・Pythonを分けると責務は明確になる一方、契約・設定・デバッグ経路も増えました。",
    next: "短期PoCではまず二層に絞り、独立して伸ばす理由ができた時点で分割します。",
  },
  {
    title: "MLはモデル以外が大きい",
    body: "前処理、入出力スキーマ、モデルのハッシュ、タイムアウトまで揃って初めて推論を運べると学びました。",
    next: "exportからAPI応答までを固定fixtureで通すスモークテストを、最初に用意します。",
  },
  {
    title: "運用設計も機能の一部",
    body: "health、readiness、run ID、構造化ログ、SLOを足すことで、動くことと観測できることの違いが見えました。",
    next: "エンドポイントが増える前に、最小限の観測項目と失敗時の見方を決めます。",
  },
  {
    title: "計画と実装を混ぜない",
    body: "短期開発ではロードマップと実装済みの境界が曖昧になりやすく、後から読み返しにくくなりました。",
    next: "READMEは現状、ADRは判断、roadmapは予定と役割を分けて更新します。",
  },
];

function ArrowIcon() {
  return (
    <svg viewBox="0 0 20 20" aria-hidden="true">
      <path d="M4 10h11M11 5l5 5-5 5" />
    </svg>
  );
}

function GitHubIcon() {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 2a10 10 0 0 0-3.16 19.49c.5.09.68-.22.68-.48v-1.87c-2.78.6-3.37-1.18-3.37-1.18-.45-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.61.07-.61 1 .07 1.53 1.03 1.53 1.03.9 1.53 2.34 1.09 2.91.83.09-.65.35-1.09.64-1.34-2.22-.25-4.56-1.11-4.56-4.94 0-1.09.39-1.98 1.03-2.68-.1-.25-.45-1.27.1-2.64 0 0 .84-.27 2.75 1.02A9.55 9.55 0 0 1 12 6.82c.85 0 1.69.11 2.49.34 1.91-1.29 2.75-1.02 2.75-1.02.55 1.37.2 2.39.1 2.64.64.7 1.03 1.59 1.03 2.68 0 3.84-2.34 4.68-4.57 4.93.36.31.68.92.68 1.86v2.76c0 .27.18.58.69.48A10 10 0 0 0 12 2Z" />
    </svg>
  );
}

function MotionTrace() {
  return (
    <svg
      className="motion-trace"
      viewBox="0 0 520 210"
      role="img"
      aria-label="モーションのキーポイントを表した抽象図"
    >
      <defs>
        <linearGradient id="trace-gradient" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stopColor="#39d3b5" />
          <stop offset="1" stopColor="#8ce8d8" />
        </linearGradient>
      </defs>
      <path
        className="trace-ghost"
        d="M26 164C91 151 117 95 176 108s69 52 130 24 74-79 188-86"
      />
      <path
        className="trace-line"
        d="M26 164C91 151 117 95 176 108s69 52 130 24 74-79 188-86"
      />
      {[
        [26, 164],
        [92, 138],
        [176, 108],
        [242, 136],
        [306, 132],
        [382, 78],
        [494, 46],
      ].map(([cx, cy], index) => (
        <g key={`${cx}-${cy}`}>
          <circle className="trace-ring" cx={cx} cy={cy} r="10" />
          <circle
            className="trace-dot"
            cx={cx}
            cy={cy}
            r={index === 4 ? "5" : "3.5"}
          />
        </g>
      ))}
    </svg>
  );
}

export default function Home() {
  return (
    <main>
      <nav className="site-nav" aria-label="メインナビゲーション">
        <a className="brand" href="#top" aria-label="Picca トップへ">
          <span className="brand-mark" aria-hidden="true">
            <span />
            <span />
            <span />
          </span>
          <span>picca</span>
        </a>

        <div className="nav-links">
          <a href="#overview">Overview</a>
          <a href="#architecture">Architecture</a>
          <a href="#learnings">Retrospective</a>
        </div>

        <a
          className="nav-github"
          href="https://github.com/Udonburo/picca"
          target="_blank"
          rel="noreferrer"
        >
          <GitHubIcon />
          <span>Repository</span>
        </a>
      </nav>

      <section className="hero" id="top">
        <div className="hero-copy">
          <div className="eyebrow">
            <span className="eyebrow-dot" />
            Archived build log · 2025
          </div>
          <h1>
            動きを読み解く、
            <br />
            <span>15秒のプロトタイプ。</span>
          </h1>
          <p className="hero-lead">
            Piccaは、身体のキーポイントから動きの特徴を捉え、
            スコアとフィードバックへ変換するモーション解析の技術検証です。
          </p>
          <p className="hero-context">
            UIからAPI・ML・運用までを横断した試行錯誤と、
            今なら変える設計判断を残した技術学習ログです。
          </p>
          <div className="hero-actions">
            <a className="button button-primary" href="#overview">
              学習記録を読む
              <ArrowIcon />
            </a>
            <a
              className="button button-secondary"
              href="https://github.com/Udonburo/picca"
              target="_blank"
              rel="noreferrer"
            >
              <GitHubIcon />
              View source
            </a>
          </div>
          <div className="archive-note">
            <span aria-hidden="true">×</span>
            開発は終了済み。実装・設計資料・振り返りをひとつの記録として保存しています。
          </div>
        </div>

        <div className="hero-visual" aria-label="Piccaの解析結果イメージ">
          <div className="visual-glow" />
          <div className="analysis-card">
            <div className="analysis-header">
              <div>
                <span className="card-label">STATIC PREVIEW</span>
                <h2>Motion analysis</h2>
              </div>
              <span className="offline-pill">
                <span /> Offline
              </span>
            </div>

            <div className="analysis-stage">
              <MotionTrace />
              <div className="stage-label stage-label-start">0s</div>
              <div className="stage-label stage-label-end">15s</div>
            </div>

            <div className="score-panel">
              <div className="score-block">
                <span>Sample score</span>
                <strong>
                  84<small>/100</small>
                </strong>
                <span className="score-caption">UI representation only</span>
              </div>
              <div className="metric-list">
                {[
                  ["Symmetry", 91],
                  ["Power", 78],
                  ["Consistency", 86],
                ].map(([label, value]) => (
                  <div className="metric" key={label}>
                    <div className="metric-label">
                      <span>{label}</span>
                      <strong>{value}</strong>
                    </div>
                    <div className="metric-track">
                      <span style={{ width: `${value}%` }} />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="float-card float-card-model">
            <span className="float-icon">◇</span>
            <div>
              <span>Inference</span>
              <strong>ONNX Runtime</strong>
            </div>
          </div>
          <div className="float-card float-card-privacy">
            <span className="float-icon">◎</span>
            <div>
              <span>Input design</span>
              <strong>Keypoints only</strong>
            </div>
          </div>
        </div>
      </section>

      <section className="project-strip" aria-label="プロジェクト概要">
        <div>
          <span>Type</span>
          <strong>Motion scoring PoC</strong>
        </div>
        <div>
          <span>Scope</span>
          <strong>Web / API / ML / Ops</strong>
        </div>
        <div>
          <span>Input</span>
          <strong>Pose keypoints</strong>
        </div>
        <div>
          <span>Status</span>
          <strong className="status-value">Archived</strong>
        </div>
      </section>

      <section className="section overview" id="overview">
        <div className="section-intro">
          <div>
            <span className="section-index">01 / OVERVIEW</span>
            <h2>小さなPoCで、<br />一通りの技術領域を。</h2>
          </div>
          <p>
            画面だけではなく、APIの境界、推論サービス、モデル配布、
            クラウド運用までをひとつの流れとして試しました。成果だけでなく、
            分割の難しさや後から見えた改善点まで振り返れる形で残しています。
          </p>
        </div>

        <div className="capability-grid">
          {capabilities.map((item) => (
            <article className="capability-card" key={item.number}>
              <div className="capability-number">{item.number}</div>
              <h3>{item.title}</h3>
              <p>{item.description}</p>
              <div className="tag-list">
                {item.tags.map((tag) => (
                  <span key={tag}>{tag}</span>
                ))}
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="section architecture" id="architecture">
        <div className="architecture-copy">
          <span className="section-index section-index-light">02 / ARCHITECTURE</span>
          <h2>入力からスコアまでを、<br />責務ごとに分ける。</h2>
          <p>
            ブラウザから送るのはポーズの座標列。Goがリクエストを受け、
            Pythonの推論サービスが前処理とスコアリングを担当する構成です。
          </p>
          <a
            href="https://github.com/Udonburo/picca/blob/main/docs/PROJECT_BRIEF.md"
            target="_blank"
            rel="noreferrer"
          >
            Project briefを見る
            <ArrowIcon />
          </a>
        </div>

        <div className="architecture-flow" aria-label="Piccaのシステム構成">
          <div className="flow-node flow-node-primary">
            <div className="flow-node-top">
              <span>01</span>
              <span className="flow-symbol">⌁</span>
            </div>
            <strong>Web client</strong>
            <small>Next.js · TypeScript</small>
          </div>
          <div className="flow-connector"><span>HTTPS / JSON</span><i /></div>
          <div className="flow-node">
            <div className="flow-node-top">
              <span>02</span>
              <span className="flow-symbol">↔</span>
            </div>
            <strong>Go gateway</strong>
            <small>Auth · routing · ops</small>
          </div>
          <div className="flow-connector"><span>Keypoint vector</span><i /></div>
          <div className="flow-node">
            <div className="flow-node-top">
              <span>03</span>
              <span className="flow-symbol">◫</span>
            </div>
            <strong>ML service</strong>
            <small>FastAPI · ONNX Runtime</small>
          </div>
          <div className="flow-output">
            <span>OUTPUT</span>
            <strong>Score · Symmetry · Power · Consistency</strong>
          </div>
        </div>
      </section>

      <section className="section learnings" id="learnings">
        <div className="learnings-heading">
          <span className="section-index">03 / RETROSPECTIVE</span>
          <h2>作ったものより、<br />次に活かす判断を。</h2>
          <a
            className="retrospective-link"
            href="https://github.com/Udonburo/picca/blob/main/docs/RETROSPECTIVE.md"
            target="_blank"
            rel="noreferrer"
          >
            振り返り全文を読む
            <ArrowIcon />
          </a>
        </div>
        <div className="learning-list">
          {learnings.map((item, index) => (
            <article className="learning-item" key={item.title}>
              <span>{String(index + 1).padStart(2, "0")}</span>
              <div>
                <h3>{item.title}</h3>
                <p>{item.body}</p>
                <div className="learning-next">
                  <span>NEXT TIME</span>
                  <p>{item.next}</p>
                </div>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="archive-cta">
        <div>
          <span className="section-index section-index-light">LEARNING ARCHIVE</span>
          <h2>開発は終了。<br />判断と反省は残す。</h2>
        </div>
        <div className="archive-cta-copy">
          <p>
            完成品としてではなく、試行錯誤・設計判断・改善点を
            読み返せる技術記録として保存しています。
          </p>
          <a
            className="button button-on-dark"
            href="https://github.com/Udonburo/picca"
            target="_blank"
            rel="noreferrer"
          >
            <GitHubIcon />
            Browse the repository
          </a>
        </div>
      </section>

      <footer>
        <a className="brand brand-footer" href="#top">
          <span className="brand-mark" aria-hidden="true">
            <span />
            <span />
            <span />
          </span>
          <span>picca</span>
        </a>
        <p>Archived prototype · Technical learning log</p>
        <span>2025</span>
      </footer>
    </main>
  );
}
