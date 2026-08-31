import { createFileRoute } from '@tanstack/react-router'
import { SyncCode, SyncGuideLayout, SyncNotice } from '../components/skills-sync-docs'

export const Route = createFileRoute('/skills-sync_/vps')({
  head: () => ({
    meta: [
      { title: 'VPS Skills Sync Setup — Code OS' },
      { name: 'description', content: 'Publish and automatically synchronize the VPS agent skill library through a private Git repository.' },
    ],
    links: [{ rel: 'canonical', href: 'https://code-os.mlvcdn.com/skills-sync/vps' }],
  }),
  component: VpsSkillsSyncPage,
})

const steps = [
  {
    title: 'Back up the VPS library',
    children: <><p>Treat the existing VPS library as the initial source of truth. Make a timestamped copy before adding Git metadata.</p><SyncCode>{`test -d ~/.agents
backup_dir="$HOME/.agents.backup.$(date +%Y%m%d-%H%M%S)"
cp -a ~/.agents "$backup_dir"
echo "Backup: $backup_dir"`}</SyncCode><SyncNotice>Nothing is deleted or moved. Keep this backup until both machines have completed several successful syncs.</SyncNotice></>,
  },
  {
    title: 'Create the private Git source',
    children: <><p>Authenticate GitHub CLI over HTTPS first. Its credential helper remains available to the background timer without an SSH agent.</p><SyncCode>{`gh auth status || gh auth login
gh config set git_protocol https
gh auth setup-git`}</SyncCode><p>Then initialize <code>~/.agents</code> directly so the repository contains the library contents—not an extra wrapper directory. Exclude credentials before the first commit.</p><SyncCode>{`cd ~/.agents
test -d .git || git init -b main
printf '%s\n' '.env' '.env.*' '*.key' '*.pem' '*token*' '*secret*' >> .gitignore
git config user.name "YOUR_NAME"
git config user.email "YOUR_EMAIL"
git add -A
git commit -m "chore: initialize shared agent skills"
gh repo create code-os-skills --private --source=. --remote=origin --push`}</SyncCode><SyncNotice security>Keep the repository private. Never include Cloudflare tokens, dashboard credentials, bypass keys, SSH keys, or machine-specific environment files.</SyncNotice><p>If the private repository already exists, replace the final command with <code>git remote add origin https://github.com/YOUR_ACCOUNT/code-os-skills.git</code> and <code>git push -u origin main</code>.</p></>,
  },
  {
    title: 'Configure and test synchronization',
    children: <><p>Open <code>/app/settings</code> and set the private GitHub repository URL, <code>~/.agents</code> checkout, and branch. The built-in worker commits local changes, rebases, and pushes. A directory lock prevents overlapping executions.</p><SyncCode>{`code-os skills-sync
git -C ~/.agents status --short --branch`}</SyncCode></>,
  },
  {
    title: 'Enable Code OS and the timer',
    children: <><p>The service installer starts Code OS now, enables it after every reboot, and creates the two-minute skills-sync timer from the same configuration. It also enables systemd user lingering so neither service depends on an interactive SSH login.</p><SyncCode>{`code-os service install
code-os doctor`}</SyncCode></>,
  },
  {
    title: 'Verify the VPS side',
    children: <><p>Trigger one execution, inspect its log, and confirm that the checkout is clean and tracking <code>origin/main</code>.</p><SyncCode>{`systemctl --user start code-os-skills-sync.service
journalctl --user -u code-os-skills-sync.service -n 30 --no-pager
git -C ~/.agents status --short --branch
systemctl --user list-timers code-os-skills-sync.timer
loginctl show-user "$USER" -p Linger`}</SyncCode><SyncNotice>A healthy run ends with “is up to date.” If Git reports a conflict, the worker stops without discarding either side.</SyncNotice></>,
  },
]

function VpsSkillsSyncPage() {
  return <SyncGuideLayout side="VPS" title="Publish the VPS skill library." description="Turn the existing ~/.agents directory into the private source shared by every development machine." steps={steps} otherSide="your computer" otherUrl="/skills-sync/local" />
}
