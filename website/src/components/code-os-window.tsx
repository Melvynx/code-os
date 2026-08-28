import { Camera, FolderGit2, GitBranch, LayoutDashboard, Play } from 'lucide-react'

const apps = [
  { name: 'lumail.io / dev', port: '3002', memory: '8.4 GB' },
  { name: 'ai-builder-club / web', port: '3100', memory: '841 MB' },
  { name: 'lumail-k9ko / dev', port: '3003', memory: '4.5 GB' },
]

export function CodeOSWindow() {
  return (
    <div className="product-window" aria-label="Code OS command center preview">
      <div className="window-bar">
        <span className="window-dot" /><span className="window-dot" /><span className="window-dot" />
        <span className="window-address">code-os.melvynx.dev</span>
      </div>
      <div className="window-body">
        <aside className="demo-sidebar">
          <div className="demo-brand"><span>S</span><b>Code OS</b></div>
          <div className="demo-nav active"><LayoutDashboard /> Overview</div>
          <div className="demo-nav"><FolderGit2 /> Projects</div>
          <div className="demo-nav"><Play /> Applications</div>
          <div className="demo-nav"><GitBranch /> Git changes</div>
          <div className="demo-nav"><Camera /> Screenshots</div>
        </aside>
        <div className="demo-main">
          <div className="demo-kicker">DEVELOPMENT ENVIRONMENT</div>
          <h3>Overview</h3>
          <div className="metric-grid">
            <div><span>Projects</span><strong>21</strong><small>discovered</small></div>
            <div><span>Worktrees</span><strong>24</strong><small>across every repo</small></div>
            <div><span>Running</span><strong>5</strong><small>Portly apps</small></div>
          </div>
          <div className="app-table">
            <div className="table-heading"><span>PORTLY · RUNNING APPLICATIONS</span><b>Live state</b></div>
            {apps.map((app) => (
              <div className="app-row" key={app.name}>
                <span><b>{app.name}</b><small>pnpm dev</small></span>
                <code>{app.port}</code><em>HEALTHY</em><span>{app.memory}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
