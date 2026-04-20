// ==================== 全局变量 ====================
const API_BASE = window.location.origin;
let activeDownloads = new Map();
let historyDownloads = [];
let currentView = 'active';

// 可选的认证 Token（如果后端启用）
// const AUTH_TOKEN = 'your-token-here';

// ==================== Toast 通知 ====================
function toast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const el = document.createElement('div');
    el.className = `toast ${type}`;
    el.textContent = message;
    container.appendChild(el);
    setTimeout(() => el.remove(), 3000);
}

// ==================== API 调用 ====================
async function apiCall(path, options = {}) {
    if (!options.headers) options.headers = {};
    if (typeof AUTH_TOKEN !== 'undefined') {
        options.headers['Authorization'] = `Bearer ${AUTH_TOKEN}`;
    }
    return fetch(`${API_BASE}${path}`, options);
}

// ==================== 格式化函数 ====================
function formatSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatSpeed(bytesPerSec) {
    if (bytesPerSec === 0) return '0 B/s';
    return formatSize(bytesPerSec) + '/s';
}

// ==================== 状态显示 ====================
function getStatusColor(status) {
    switch(status) {
        case 'downloading': return 'downloading';
        case 'paused': return 'paused';
        case 'completed': return 'completed';
        case 'error': return 'error';
        default: return 'downloading';
    }
}

function getStatusLabel(status) {
    switch(status) {
        case 'downloading': return '下载中';
        case 'paused': return '已暂停';
        case 'completed': return '已完成';
        case 'error': return '错误';
        default: return status;
    }
}

// ==================== 数据获取 ====================
async function fetchDownloads() {
    try {
        const response = await apiCall('/api/downloads');
        if (response.ok) {
            const downloads = await response.json();
            activeDownloads.clear();
            downloads.forEach(dl => activeDownloads.set(dl.id, dl));
            renderList();
            updateStats();
            
            // 更新连接状态
            const indicator = document.getElementById('statusIndicator');
            indicator.className = 'status-indicator connected';
            indicator.querySelector('.status-text').textContent = '已连接';
        } else {
            console.error('Fetch downloads failed:', response.status);
            const indicator = document.getElementById('statusIndicator');
            indicator.className = 'status-indicator disconnected';
            indicator.querySelector('.status-text').textContent = 
                response.status === 401 ? '认证失败' : `服务器错误 (${response.status})`;
        }
    } catch (err) {
        console.error('Failed to fetch downloads:', err);
        const indicator = document.getElementById('statusIndicator');
        indicator.className = 'status-indicator disconnected';
        indicator.querySelector('.status-text').textContent = '连接断开';
    }
}

async function fetchHistory() {
    try {
        const response = await apiCall('/api/downloads/history');
        if (response.ok) {
            historyDownloads = await response.json();
            if (currentView === 'history') renderList();
            updateStats();
        }
    } catch (err) {
        console.error('Failed to fetch history:', err);
    }
}

// ==================== 统计更新 ====================
function updateStats() {
    let active = 0;
    let totalSpeed = 0;
    
    activeDownloads.forEach(dl => {
        if (dl.status === 'downloading') {
            active++;
            totalSpeed += dl.speed || 0;
        }
    });
    
    document.getElementById('activeCount').textContent = active;
    document.getElementById('totalSpeed').textContent = formatSpeed(totalSpeed);
    document.getElementById('completedCount').textContent = historyDownloads.length;
}

// ==================== 列表渲染 ====================
function renderList() {
    const container = document.getElementById('downloadList');
    const list = currentView === 'active' 
        ? Array.from(activeDownloads.values()).filter(dl => dl.status !== 'completed') 
        : historyDownloads;
    
    if (list.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <p>${currentView === 'active' ? '暂无进行中的下载' : '暂无历史记录'}</p>
            </div>
        `;
        return;
    }

    container.innerHTML = '';
    list.forEach(dl => {
        const totalSize = dl.total_size || 0;
        const downloaded = dl.downloaded || 0;
        const progress = dl.progress !== undefined ? dl.progress.toFixed(1) : 
            (totalSize > 0 ? (downloaded / totalSize * 100).toFixed(1) : 0);
        
        const item = document.createElement('div');
        item.className = 'download-item';
        
        // 构建操作按钮
        let actions = '';
        if (currentView === 'active') {
            if (dl.status === 'paused') {
                actions = `
                    <button onclick="resumeDownload('${dl.id}')" class="action-btn resume" title="恢复">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                        </svg>
                    </button>
                `;
            } else if (dl.status === 'downloading') {
                actions = `
                    <button onclick="pauseDownload('${dl.id}')" class="action-btn pause" title="暂停">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6" />
                        </svg>
                    </button>
                `;
            }
            actions += `
                <button onclick="deleteDownload('${dl.id}')" class="action-btn delete" title="删除">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-4v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                </button>
            `;
        } else {
            if (dl.file_exists) {
                actions = `
                    <button onclick="openFile('${dl.id}')" class="btn btn-sm btn-primary">打开文件</button>
                    <button onclick="openFolder('${dl.id}')" class="btn btn-sm btn-secondary">打开目录</button>
                `;
            } else {
                actions = `
                    <span class="text-xs text-red-500">文件已丢失</span>
                    <button onclick="redownload('${dl.id}')" class="btn btn-sm btn-primary">重新下载</button>
                `;
            }
        }
        
        item.innerHTML = `
            <div class="download-header">
                <div class="download-info">
                    <h3 class="download-filename" title="${dl.filename || dl.id}">${dl.filename || dl.id}</h3>
                    <p class="download-url">${dl.url}</p>
                </div>
                <div class="download-actions">
                    ${actions}
                </div>
            </div>
            <div class="progress-container">
                <div class="progress-bar-bg">
                    <div class="progress-bar" style="width: ${progress}%"></div>
                </div>
                <span class="progress-text">${progress}%</span>
            </div>
            <div class="download-meta">
                <div class="download-stats">
                    <span>${formatSize(downloaded)} / ${formatSize(totalSize)}</span>
                    ${dl.speed > 0 ? `<span class="download-speed">${formatSpeed(dl.speed)}</span>` : ''}
                    ${dl.eta > 0 ? `<span>剩余 ${dl.eta}s</span>` : ''}
                </div>
                <span class="status-badge ${getStatusColor(dl.status)}">${getStatusLabel(dl.status)}</span>
            </div>
        `;
        
        container.appendChild(item);
    });
}

// ==================== 下载操作 ====================
async function pauseDownload(id) {
    await apiCall(`/api/downloads/${id}/pause`, { method: 'POST' });
    fetchDownloads();
}

async function resumeDownload(id) {
    await apiCall(`/api/downloads/${id}/resume`, { method: 'POST' });
    fetchDownloads();
}

async function deleteDownload(id) {
    if (!confirm('确定要删除这个下载任务吗？')) return;
    
    await apiCall(`/api/downloads/${id}`, { method: 'DELETE' });
    fetchDownloads();
    toast('已删除任务', 'success');
}

async function openFile(id) {
    await apiCall(`/api/files/${id}/open`, { method: 'POST' });
}

async function openFolder(id) {
    await apiCall(`/api/files/${id}/folder`, { method: 'POST' });
}

async function redownload(id) {
    try {
        const response = await apiCall(`/api/downloads/${id}/redownload`, { method: 'POST' });
        if (response.ok) {
            document.getElementById('showActive').click();
            fetchDownloads();
            toast('已重新开始下载', 'success');
        } else {
            toast('重新下载失败: ' + await response.text(), 'error');
        }
    } catch (err) {
        toast('重新下载出错: ' + err.message, 'error');
    }
}

// ==================== 添加下载 ====================
document.getElementById('addDownloadForm').onsubmit = async (e) => {
    e.preventDefault();
    const url = document.getElementById('urlInput').value.trim();
    if (!url) return;

    try {
        const response = await apiCall('/api/downloads', { 
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url: url })
        });
        if (response.ok) {
            document.getElementById('urlInput').value = '';
            fetchDownloads();
            toast('已添加下载任务', 'success');
        } else {
            toast('添加失败: ' + await response.text(), 'error');
        }
    } catch (err) {
        toast('添加出错: ' + err.message, 'error');
    }
};

// ==================== 视图切换 ====================
document.getElementById('showActive').onclick = () => {
    currentView = 'active';
    document.getElementById('showActive').classList.add('active');
    document.getElementById('showHistory').classList.remove('active');
    document.getElementById('clearHistoryBtn').classList.add('hidden');
    renderList();
};

document.getElementById('showHistory').onclick = () => {
    currentView = 'history';
    document.getElementById('showHistory').classList.add('active');
    document.getElementById('showActive').classList.remove('active');
    document.getElementById('clearHistoryBtn').classList.remove('hidden');
    renderList();
};

document.getElementById('clearHistoryBtn').onclick = async () => {
    if (!confirm('确定要清空所有历史记录吗？')) return;
    
    try {
        const response = await apiCall('/api/downloads/history', { method: 'DELETE' });
        if (response.ok) {
            historyDownloads = [];
            renderList();
            toast('历史记录已清空', 'success');
        }
    } catch (err) {
        toast('清空失败: ' + err.message, 'error');
    }
};

// ==================== 设置模态框 ====================
const settingsModal = document.getElementById('settingsModal');
const settingsBtn = document.getElementById('settingsBtn');
const closeSettings = document.getElementById('closeSettings');
const cancelSettings = document.getElementById('cancelSettings');

settingsBtn.onclick = () => {
    settingsModal.classList.remove('hidden');
    loadSettings();
};

closeSettings.onclick = () => settingsModal.classList.add('hidden');
cancelSettings.onclick = () => settingsModal.classList.add('hidden');

settingsModal.onclick = (e) => {
    if (e.target === settingsModal) settingsModal.classList.add('hidden');
};

async function loadSettings() {
    try {
        const response = await apiCall('/api/settings');
        if (response.ok) {
            const settings = await response.json();
            // 填充设置表单
            Object.keys(settings).forEach(key => {
                const el = document.getElementById('setting-' + key);
                if (el) el.value = settings[key];
            });
        }
    } catch (err) {
        console.error('Failed to load settings:', err);
    }
}

document.getElementById('settingsForm').onsubmit = async (e) => {
    e.preventDefault();
    
    const updatedSettings = {
        max_downloads: parseInt(document.getElementById('setting-max_downloads').value),
        default_dir: document.getElementById('setting-default_dir').value,
        proxy: document.getElementById('setting-proxy').value,
        timeout: parseInt(document.getElementById('setting-timeout').value),
    };

    try {
        const response = await apiCall('/api/settings', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(updatedSettings)
        });
        if (response.ok) {
            settingsModal.classList.add('hidden');
            toast('设置已保存', 'success');
        } else {
            toast('保存失败: ' + await response.text(), 'error');
        }
    } catch (err) {
        toast('保存出错: ' + err.message, 'error');
    }
};

// ==================== SSE 实时更新 ====================
function setupSSE() {
    let url = `${API_BASE}/api/events`;
    if (typeof AUTH_TOKEN !== 'undefined') {
        url += `?token=${AUTH_TOKEN}`;
    }
    
    const evSource = new EventSource(url);
    
    evSource.onmessage = (event) => {
        // 收到事件后刷新列表
        fetchDownloads();
    };
    
    evSource.onerror = () => {
        evSource.close();
        setTimeout(setupSSE, 5000); // 5秒后重试
    };
}

// ==================== 初始化 ====================
setInterval(fetchDownloads, 1000);  // 每秒刷新下载列表
setInterval(fetchHistory, 5000);    // 每5秒刷新历史记录

fetchDownloads();
fetchHistory();
setupSSE();
