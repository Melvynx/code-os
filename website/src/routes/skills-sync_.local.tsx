import { createFileRoute } from '@tanstack/react-router'
import { SyncCode, SyncGuideLayout, SyncNotice } from '../components/skills-sync-docs'

export const Route = createFileRoute('/skills-sync_/local')({
  head: () => ({
    meta: [
      { title: 'Computer Skills Sync Setup — StackEnv' },
      { name: 'description', content: 'Connect a Mac or local development computer to the shared private agent skill library.' },
    ],
    links: [{ rel: 'canonical', href: 'https://stackend.codelynx.dev/skills-sync/local' }],
  }),
  component: LocalSkillsSyncPage,
})

const steps = [
  {
    title: 'Prepare Git authentication',
    children: <><p>Make sure your computer can read and write the private repository without an interactive password prompt. On macOS, keep the SSH key in Keychain.</p><SyncCode>{`test ! -f ~/.ssh/id_ed25519 || ssh-add --apple-use-keychain ~/.ssh/id_ed25519
ssh -T git@github.com || true
git config --global user.name "YOUR_NAME"
git config --global user.email "YOUR_EMAIL"`}</SyncCode><p>If you use HTTPS instead, authenticate with <code>gh auth login</code>; GitHub CLI stores the credential in the macOS Keychain.</p></>,
  },
  {
    title: 'Preserve the local library',
    children: <><p>Move the existing local directory aside, then clone the VPS-backed repository into the canonical <code>~/.agents</code> path.</p><SyncCode>{`backup_dir="$HOME/.agents.backup.$(date +%Y%m%d-%H%M%S)"
if [ -d ~/.agents ]; then mv ~/.agents "$backup_dir"; fi
git clone git@github.com:YOUR_ACCOUNT/stackenv-skills.git ~/.agents
echo "Previous local skills: $backup_dir"`}</SyncCode><SyncNotice>Do not delete the backup. Compare it with the cloned library and manually copy only the local skills or rules you still need.</SyncNotice></>,
  },
  {
    title: 'Install the same worker',
    children: <><p>Both machines use the identical audited script and private Git branch. Run it once interactively before scheduling it.</p><SyncCode>{`mkdir -p ~/.local/bin
curl -fsSL https://stackend.codelynx.dev/skills-sync.sh -o ~/.local/bin/stackenv-skills-sync
chmod 700 ~/.local/bin/stackenv-skills-sync
~/.local/bin/stackenv-skills-sync`}</SyncCode></>,
  },
  {
    title: 'Schedule it with launchd',
    children: <><p>The LaunchAgent starts after login and runs every two minutes. Logs stay in <code>~/Library/Logs/StackEnv</code>.</p><SyncCode>{`mkdir -p ~/Library/LaunchAgents ~/Library/Logs/StackEnv
cat > ~/Library/LaunchAgents/dev.stackenv.skills-sync.plist <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>dev.stackenv.skills-sync</string>
  <key>ProgramArguments</key><array>
    <string>/bin/zsh</string><string>-lc</string>
    <string>$HOME/.local/bin/stackenv-skills-sync</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>StartInterval</key><integer>120</integer>
  <key>StandardOutPath</key><string>LOCAL_LOG_PATH/stackenv-skills-sync.log</string>
  <key>StandardErrorPath</key><string>LOCAL_LOG_PATH/stackenv-skills-sync.error.log</string>
</dict></plist>
EOF
sed -i '' "s|LOCAL_LOG_PATH|$HOME/Library/Logs/StackEnv|g" ~/Library/LaunchAgents/dev.stackenv.skills-sync.plist
launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/dev.stackenv.skills-sync.plist 2>/dev/null || true
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/dev.stackenv.skills-sync.plist`}</SyncCode></>,
  },
  {
    title: 'Verify both directions',
    children: <><p>Create a harmless test file on the computer, run a sync, then confirm it arrives on the VPS after the next timer execution.</p><SyncCode>{`touch ~/.agents/.sync-check
~/.local/bin/stackenv-skills-sync
launchctl kickstart -k gui/$(id -u)/dev.stackenv.skills-sync
tail -n 30 ~/Library/Logs/StackEnv/stackenv-skills-sync.log
git -C ~/.agents status --short --branch`}</SyncCode><p>On the VPS, run <code>systemctl --user start stackenv-skills-sync.service</code>, check that <code>.sync-check</code> exists, then remove it and sync once more.</p><SyncNotice security>Only <code>~/.agents</code> is synchronized. Project repositories, StackEnv configuration, tokens, and screenshot bypass keys stay machine-local.</SyncNotice></>,
  },
]

function LocalSkillsSyncPage() {
  return <SyncGuideLayout side="YOUR COMPUTER · macOS" title="Connect your local skill library." description="Clone the VPS-backed source, preserve any local-only work, and keep both sides synchronized in the background." steps={steps} otherSide="the VPS" otherUrl="/skills-sync/vps" />
}
