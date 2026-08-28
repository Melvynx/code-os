const state = { snapshot: null, query: "", section: "overview" };
const $ = (selector) => document.querySelector(selector);
const escapeHTML = (value = "") => String(value).replace(/[&<>'"]/g, char => ({"&":"&amp;","<":"&lt;",">":"&gt;","'":"&#39;",'"':"&quot;"})[char]);
const formatBytes = value => {
  if (!value) return "—";
  const units = ["B", "KB", "MB", "GB"];
  let index = 0, number = value;
  while (number >= 1024 && index < units.length - 1) { number /= 1024; index++; }
  return `${number.toFixed(index > 1 ? 1 : 0)} ${units[index]}`;
};
const ago = value => {
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(value).getTime()) / 1000));
  if (seconds < 60) return `${seconds}s ago`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  return `${Math.floor(seconds / 86400)}d ago`;
};
const matches = (...values) => !state.query || values.some(value => String(value || "").toLowerCase().includes(state.query));
const changeCount = project => Object.entries(project.git).reduce((sum, [key, value]) => ["modified","added","deleted","untracked","conflicts"].includes(key) ? sum + value : sum, 0);
const empty = message => `<div class="empty">${escapeHTML(message)}</div>`;
const safeHref = value => {
  if (!value) return "";
  try {
    const parsed = new URL(value, location.origin);
    return ["http:", "https:"].includes(parsed.protocol) ? escapeHTML(parsed.href) : "";
  } catch { return ""; }
};

async function request(path, options) {
  const response = await fetch(path, options);
  if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
  return response.json();
}

async function load(force = false) {
  const button = $("#refresh");
  button.disabled = true;
  button.textContent = force ? "Refreshing…" : "Loading…";
  try {
    state.snapshot = await request(force ? "/api/refresh" : "/api/snapshot", force ? { method: "POST" } : undefined);
    render();
    toast(force ? "Environment refreshed" : "Command center connected");
  } catch (error) {
    toast(`Could not load environment: ${error.message}`);
  } finally {
    button.disabled = false;
    button.textContent = "Refresh";
  }
}

function render() {
  if (!state.snapshot) return;
  renderIdentity();
  renderMetrics();
  renderApps();
  renderGit();
  renderProjects();
  renderScreenshots();
}

function renderIdentity() {
  const generated = state.snapshot.generatedAt;
  $("#environment-name").textContent = location.hostname === "127.0.0.1" ? "StackEnv local" : location.hostname;
  $("#last-refresh").textContent = generated ? `Updated ${ago(generated)}` : "Not scanned";
}

function renderMetrics() {
  const projects = state.snapshot.projects || [];
  const apps = state.snapshot.applications || [];
  const screenshots = state.snapshot.screenshots || [];
  const metrics = [
    [projects.length, "Projects", "discovered"],
    [projects.filter(project => changeCount(project) > 0).length, "Modified", "worktrees"],
    [apps.filter(app => app.state === "running").length, "Running", "Portly apps"],
    [screenshots.length, "Screenshots", "indexed"],
    [(state.snapshot.warnings || []).length, "Warnings", "diagnostics"],
  ];
  $("#metrics").innerHTML = metrics.map(([value, label, hint]) => `<article class="metric"><small>${label}</small><strong>${value}</strong><em>${hint}</em></article>`).join("");
}

function appRow(app, full = false) {
  const healthy = app.healthy === true;
  const statusClass = app.state !== "running" ? "muted" : healthy ? "" : "warning";
  const link = safeHref(app.publicUrl || app.url);
  return `<div class="table-row"><div><strong>${escapeHTML(app.projectName)} / ${escapeHTML(app.name)}</strong><small>${escapeHTML(app.command)}</small></div><span>${app.port || "—"}</span><span class="badge ${statusClass}">${healthy ? "Healthy" : escapeHTML(app.state)}</span><span>${formatBytes(app.residentMemoryBytes || app.memoryBytes)}</span>${full ? `<span>${link ? `<a href="${link}" target="_blank" rel="noreferrer">Open ↗</a>` : "—"}</span>` : ""}</div>`;
}

function renderApps() {
  const apps = (state.snapshot.applications || []).filter(app => matches(app.name, app.projectName, app.command, app.port));
  $("#overview-apps").innerHTML = apps.slice(0, 5).map(app => appRow(app)).join("") || empty("No managed applications");
  $("#apps-table").innerHTML = apps.map(app => appRow(app, true)).join("") || empty("No matching applications");
}

function renderGit() {
  const projects = (state.snapshot.projects || []).filter(project => changeCount(project) > 0 && matches(project.name, project.path, project.git.branch));
  $("#overview-changes").innerHTML = projects.slice(0, 5).map(project => `<div class="stack-item"><div><strong>${escapeHTML(project.name)}</strong><small>${escapeHTML(project.git.branch || "detached")}</small></div><span class="change-count">${changeCount(project)} changes</span></div>`).join("") || empty("Every repository is clean");
  $("#git-grid").innerHTML = projects.map(project => `<article class="card"><header><div><h3>${escapeHTML(project.name)}</h3><p>${escapeHTML(project.path)}</p></div><span class="badge warning">${changeCount(project)} changes</span></header><div class="card-stats"><div><strong>${project.git.modified}</strong><small>Modified</small></div><div><strong>${project.git.untracked}</strong><small>Untracked</small></div><div><strong>${project.git.conflicts}</strong><small>Conflicts</small></div></div><div class="chips"><span class="chip">${escapeHTML(project.git.branch || "detached")}</span>${project.git.ahead ? `<span class="chip">↑ ${project.git.ahead}</span>` : ""}${project.git.behind ? `<span class="chip">↓ ${project.git.behind}</span>` : ""}</div></article>`).join("") || empty("No matching Git changes");
}

function renderProjects() {
  const projects = (state.snapshot.projects || []).filter(project => matches(project.name, project.path, project.git.branch, ...(project.subprojects || []).map(item => item.name)));
  $("#projects-grid").innerHTML = projects.map(project => `<article class="card"><header><div><h3>${escapeHTML(project.name)}</h3><p>${escapeHTML(project.path)}</p></div><span class="badge ${changeCount(project) ? "warning" : ""}">${changeCount(project) ? `${changeCount(project)} changes` : "Clean"}</span></header><div class="card-stats"><div><strong>${escapeHTML(project.git.branch || "—")}</strong><small>Branch</small></div><div><strong>${(project.subprojects || []).length}</strong><small>Subprojects</small></div></div><div class="chips">${(project.subprojects || []).slice(0, 8).map(item => `<span class="chip">${escapeHTML(item.name)} · ${escapeHTML(item.kind)}</span>`).join("") || `<span class="chip">Repository root</span>`}</div></article>`).join("") || empty("No matching projects");
}

function renderScreenshots() {
  const screenshots = (state.snapshot.screenshots || []).filter(image => matches(image.name, image.project, image.group));
  $("#overview-screenshots").innerHTML = screenshots.slice(0, 6).map(image => `<a href="${safeHref(image.url)}" target="_blank" rel="noreferrer" title="${escapeHTML(image.name)}"><img src="${safeHref(image.url)}" alt="${escapeHTML(image.name)}" loading="lazy"></a>`).join("") || empty("No screenshots indexed");
  $("#screenshots-grid").innerHTML = screenshots.map(image => `<article class="gallery-card"><a class="gallery-image" href="${safeHref(image.url)}" target="_blank" rel="noreferrer"><img src="${safeHref(image.url)}" alt="${escapeHTML(image.name)}" loading="lazy"></a><div class="gallery-meta"><strong>${escapeHTML(image.name)}</strong><small>${escapeHTML(image.project || image.group || "Screenshot")} · ${ago(image.createdAt)}</small></div></article>`).join("") || empty("No matching screenshots");
}

function navigate(section) {
  state.section = document.getElementById(section) ? section : "overview";
  document.querySelectorAll(".view").forEach(view => view.classList.toggle("active", view.id === state.section));
  document.querySelectorAll(".nav-item").forEach(item => item.classList.toggle("active", item.dataset.section === state.section));
  $("#page-title").textContent = document.querySelector(`[data-section="${state.section}"]`)?.textContent.trim() || "Overview";
}

let toastTimer;
function toast(message) {
  const element = $("#toast");
  element.textContent = message;
  element.classList.add("show");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => element.classList.remove("show"), 2800);
}

$("#refresh").addEventListener("click", () => load(true));
$("#search").addEventListener("input", event => { state.query = event.target.value.trim().toLowerCase(); render(); });
window.addEventListener("hashchange", () => navigate(location.hash.slice(1)));
document.querySelectorAll("[data-jump]").forEach(link => link.addEventListener("click", () => navigate(link.dataset.jump)));
navigate(location.hash.slice(1));
load();
setInterval(() => load(), 30000);
