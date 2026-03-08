<script>
  import {
    DownloadPlaylist,
    SelectFolder,
    UpdateConfig,
    GetConfig,
  } from "../wailsjs/go/main/App";
  import { models } from "../wailsjs/go/models"; // Import the models namespace
  import { onMount } from "svelte";
  import { EventsOn } from "../wailsjs/runtime/runtime"; // Import Wails events

  let activeDownloads = [];

  let showSettings = false;
  let cfg = new models.Config({
    download_path: "",
    workers: 3,
    debug_mode: false,
    light_mode: false,
    retries: [],
  });

  let downloadPath = "No folder selected";

  let url = "";
  let status = "Waiting for URL...";
  let processing = false;

  // Total progress state
  let totalTracks = 0;
  let startTime = null;
  let elapsed = "00:00:00";
  let timer = null;

  $: completedCount = activeDownloads.filter(
    (t) => t.status === "Completed",
  ).length;
  $: failedCount = activeDownloads.filter((t) => t.status === "Failed").length;
  $: skippedCount = activeDownloads.filter(
    (t) => t.status === "Skipped",
  ).length;
  $: percent =
    totalTracks > 0 ? Math.round((completedCount / totalTracks) * 100) : 0;
  let downloadsCollapsed = false;

  function toggleDownloads() {
    downloadsCollapsed = !downloadsCollapsed;
  }

  onMount(() => {
    EventsOn("download_progress", (data) => {
      const index = activeDownloads.findIndex((t) => t.id === data.id);

      if (index === -1) {
        // Create the new div entry
        activeDownloads = [...activeDownloads, data];
      } else {
        // Update the existing div entry (only update the fields Go sent)
        activeDownloads[index] = { ...activeDownloads[index], ...data };
        activeDownloads = [...activeDownloads];
      }
    });
  });

  function pad(n) {
    return String(n).padStart(2, "0");
  }

  function applyTheme(light) {
    try {
      if (light) document.body.classList.add("light-theme");
      else document.body.classList.remove("light-theme");
    } catch (e) {
      /* ignore in non-browser contexts */
    }
  }

  async function startDownload() {
    if (!url) return;
    // Reset total progress and per-download state for a new job
    totalTracks = 0;
    startTime = null;
    elapsed = "00:00:00";
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
    // Clear visible list so new queued events repopulate it
    activeDownloads = [];

    processing = true;
    status = "Contacting backend...";

    // Calls the Go function
    const response = await DownloadPlaylist(url);
    status = response;

    // Parse "Queued N tracks" from backend response to set total
    const m =
      response && response.match && response.match(/Queued\s+(\d+)\s+tracks/);
    if (m) {
      totalTracks = parseInt(m[1], 10);
      startTime = Date.now();
      if (timer) clearInterval(timer);
      timer = setInterval(() => {
        if (!startTime) return;
        const ms = Date.now() - startTime;
        const hrs = Math.floor(ms / 3600000);
        const mins = Math.floor((ms % 3600000) / 60000);
        const secs = Math.floor((ms % 60000) / 1000);
        elapsed = `${pad(hrs)}:${pad(mins)}:${pad(secs)}`;

        if (
          totalTracks > 0 &&
          completedCount + failedCount + skippedCount >= totalTracks
        ) {
          clearInterval(timer);
          timer = null;
        }
      }, 1000);
    }
    processing = false;
  }

  async function handleBrowse() {
    const folder = await SelectFolder();
    if (folder) {
      downloadPath = folder;
    }
  }

  onMount(async () => {
    const savedCfg = await GetConfig();
    if (savedCfg) {
      cfg = savedCfg;
      // Use config value for visible download path when the app opens
      if (savedCfg.download_path) downloadPath = savedCfg.download_path;
      // Apply theme from config
      if (typeof savedCfg.light_mode !== 'undefined') applyTheme(savedCfg.light_mode);
    }
  });

  async function pickFolder() {
    const path = await SelectFolder();
    if (path) cfg.download_path = path;
  }

  async function save() {
    const status = await UpdateConfig(cfg);
    // apply immediately
    applyTheme(cfg.light_mode);
    showSettings = false;
  }
</script>

{#if showSettings}
  <div class="modal-overlay">
    <div class="settings-card">
      <h3>Application Settings</h3>

      <div class="field">
        <label for="download-path">Download Location:</label>
        <div class="row">
          <input
            id="download-path"
            type="text"
            bind:value={cfg.download_path}
            readonly
          />
          <button id="change-dir" on:click={pickFolder}>Browse</button>
        </div>
      </div>

      <div class="field">
        <label for="worker-range">Concurrent Workers ({cfg.workers}):</label>
        <input
          id="worker-range"
          type="range"
          bind:value={cfg.workers}
          min="1"
          max="5"
        />
      </div>

      <div class="field">
        <label for="light-mode">Light Mode:</label>
        <div class="row">
          <input id="light-mode" type="checkbox" bind:checked={cfg.light_mode} on:change={() => applyTheme(cfg.light_mode)} />
          <span style="align-self:center;margin-left:8px">{cfg.light_mode ? 'Enabled' : 'Disabled'}</span>
        </div>
      </div>

      <div class="actions">
        <button on:click={() => (showSettings = false)}>Cancel</button>
        <button class="primary" on:click={save}>Save Changes</button>
      </div>
    </div>
  </div>
{/if}

<body>
  <div class="top-bar">
    <span id="logo">GoDown</span>
    <button id="settings-icon" on:click={() => (showSettings = true)}>
      <span class="mdi-light--settings"></span>
    </button>
  </div>
  <div class="input-box">
    <input
      id="url-input"
      bind:value={url}
      placeholder="Paste URL Here..."
      type="text"
    />
    <button id="download-btn" disabled={processing} on:click={startDownload}>
      <span>{processing ? "Processing..." : "Download"}</span>
    </button>
  </div>
  <div class="dir-box">
    <p id="path"><strong>{downloadPath}</strong></p>
    <button id="change-dir" on:click={handleBrowse}>Change Folder</button>
  </div>

  <div class="active-downloads-section">
    <div class="active-downloads-header">
      <span>Active Downloads ({activeDownloads.length})</span>
      <button
        class="toggle-btn"
        on:click={toggleDownloads}
        aria-pressed={downloadsCollapsed}
      >
        {downloadsCollapsed ? "Show" : "Hide"}
      </button>
    </div>

    {#if !downloadsCollapsed}
      <div class="downloads-container">
        {#each activeDownloads as track (track.id || track.title)}
          <div
            class="download-card {track.status
              .toLowerCase()
              .replace('...', '')}"
          >
            <div class="info">
              <span class="track-name">{track.title}</span>
              <span class="track-status">{track.status}</span>
            </div>
          </div>
        {:else}
          <div class="empty-state">No active downloads</div>
        {/each}
      </div>
    {/if}
  </div>
  <div class="total-progress-container" aria-live="polite">
    <div class="total-progress-header">
      <div class="left">
        Downloaded {completedCount} of {totalTracks || activeDownloads.length}
      </div>
      <div class="right">{percent}%</div>
    </div>
    <div class="total-progress-bar">
      <div class="total-progress-fill" style="width: {percent}%"></div>
    </div>
    <div class="total-progress-meta">
      <span>Time: {elapsed}</span>
      <span>Skipped: {skippedCount}</span>
      <span>Failed: {failedCount}</span>
    </div>
  </div>

  <div class="status">
    <p class="status-text">{status}</p>
  </div>
</body>
