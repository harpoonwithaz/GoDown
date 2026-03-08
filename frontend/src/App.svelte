<script>
  import { DownloadPlaylist, SelectFolder, UpdateConfig, GetConfig } from '../wailsjs/go/main/App';
  import { models } from '../wailsjs/go/models'; // Import the models namespace
  import { onMount } from 'svelte';

  let url = "";
  let status = "Waiting for URL...";
  let processing = false;

  let downloadPath = "No folder selected";

  let showSettings = false;
  let cfg = new models.Config({
    download_path: "",
    workers: 3,
    debug_mode: false,
    retries: []
  });

  async function startDownload() {
    if (!url) return;
    processing = true;
    status = "Contacting backend...";
    
    // Calls the Go function
    const response = await DownloadPlaylist(url);
    status = response;
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
    if (savedCfg) cfg = savedCfg;
  });

  async function pickFolder() {
    const path = await SelectFolder();
    if (path) cfg.download_path = path;
  }

  async function save() {
    const status = await UpdateConfig(cfg);
    alert(status);
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
        <button on:click={pickFolder}>Browse</button>
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

    <div class="actions">
      <button on:click={() => showSettings = false}>Cancel</button>
      <button class="primary" on:click={save}>Save Changes</button>
    </div>
  </div>
</div>
{/if}

<body>
  <div class="top-bar">
    <span id="logo">GoDown</span>
    <button id="settings-icon" on:click={() => showSettings = true}>
      <span class="mdi-light--settings"></span>
    </button>
    
  </div>
  <div class="input-box">
    <input id="url-input" bind:value={url} placeholder="Paste Playlist URL Here..." type="text" />
    <button id="download-btn" disabled={processing} on:click={startDownload}>
      {processing ? "Processing..." : "Download"}
    </button>
  </div>
  <div class="dir-box">
    <p id="path"><strong>{downloadPath}</strong></p>
    <button id="change-dir" on:click={handleBrowse}>Change Folder</button>
  </div>
  <div class="active-downloads">
    <span>Active Downloads</span>
  </div>
  <div class="status">
    <p class="status-text">{status}</p>
  </div>
</body>
