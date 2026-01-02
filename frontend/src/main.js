import './style.css';
import {SelectImages, SelectOutputDir, CompressImage, GetImageInfo, CreateGifFromSequence} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

// 状态管理
let state = {
    mode: 'compress',    // 'compress' 或 'gif'
    files: [],           // {path, name, size, width, height, format, preview, status, result}
    outputDir: '',
    options: {
        quality: 80,
        maxWidth: 0,
        maxHeight: 0,
        outputFormat: 'original',
        keepAspect: true
    },
    gifOptions: {
        frameDelay: 100,  // 毫秒
        loopCount: 0,     // 0=无限循环
        maxWidth: 0,
        maxHeight: 0,
        outputName: 'animation'
    },
    isProcessing: false,
    stopRequested: false,  // 停止压缩请求
    currentIndex: -1,
    totalSaved: 0
};

// 初始化界面
function initUI() {
    document.querySelector('#app').innerHTML = `
        <div class="app-container">
            <!-- 顶部工具栏 -->
            <div class="toolbar">
                <div class="toolbar-left">
                    <!-- 模式切换 -->
                    <div class="mode-switch">
                        <button class="mode-btn active" data-mode="compress" title="图片压缩">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <polyline points="16 16 12 12 8 16"/>
                                <line x1="12" y1="12" x2="12" y2="21"/>
                                <path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"/>
                            </svg>
                            压缩
                        </button>
                        <button class="mode-btn" data-mode="gif" title="序列帧转GIF">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/>
                                <line x1="7" y1="2" x2="7" y2="22"/>
                                <line x1="17" y1="2" x2="17" y2="22"/>
                                <line x1="2" y1="12" x2="22" y2="12"/>
                            </svg>
                            GIF
                        </button>
                    </div>
                    <div class="toolbar-divider"></div>
                    <button class="btn btn-icon" id="addFiles" title="添加图片">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                            <polyline points="17 8 12 3 7 8"/>
                            <line x1="12" y1="3" x2="12" y2="15"/>
                        </svg>
                        添加图片
                    </button>
                    <button class="btn btn-icon" id="selectOutput" title="输出目录">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
                        </svg>
                        输出目录
                    </button>
                    <button class="btn btn-icon btn-clear" id="clearAll" title="清空列表">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="3 6 5 6 21 6"/>
                            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                        </svg>
                    </button>
                </div>
                <div class="toolbar-right">
                    <span class="output-path" id="outputPath">未选择输出目录</span>
                </div>
            </div>

            <!-- 主内容区 -->
            <div class="main-content">
                <!-- 左侧文件列表 -->
                <div class="file-panel">
                    <div class="drop-zone" id="dropZone">
                        <div class="drop-zone-content" id="dropZoneContent">
                            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                <polyline points="17 8 12 3 7 8"/>
                                <line x1="12" y1="3" x2="12" y2="15"/>
                            </svg>
                            <p id="dropZoneTitle">拖放图片到这里</p>
                            <span id="dropZoneSubtitle">或点击上方按钮选择文件</span>
                            <span class="format-hint">支持 JPG, PNG, GIF, WebP, TIFF</span>
                        </div>
                        <div class="file-list" id="fileList"></div>
                    </div>
                </div>

                <!-- 右侧设置面板 -->
                <div class="settings-panel">
                    <!-- 压缩模式设置 -->
                    <div id="compressSettings">
                        <div class="settings-section">
                            <h3>输出格式</h3>
                            <div class="format-buttons" id="formatButtons">
                                <button class="format-btn active" data-format="original">原格式</button>
                                <button class="format-btn" data-format="jpeg">JPEG</button>
                                <button class="format-btn" data-format="png">PNG</button>
                                <button class="format-btn" data-format="webp">WebP</button>
                            </div>
                        </div>

                        <div class="settings-section">
                            <h3>压缩质量</h3>
                            <div class="quality-slider">
                                <input type="range" id="qualitySlider" min="1" max="100" value="80">
                                <div class="quality-markers">
                                    <span class="marker" style="left: 0%">模糊</span>
                                    <span class="marker" style="left: 50%">适中</span>
                                    <span class="marker lossless" style="left: 100%">无损</span>
                                </div>
                                <div class="quality-labels">
                                    <span>低质量</span>
                                    <span class="quality-value" id="qualityValue">80%</span>
                                    <span>高质量</span>
                                </div>
                                <div class="quality-hint" id="qualityHint"></div>
                            </div>
                        </div>

                        <div class="settings-section">
                            <h3>尺寸限制</h3>
                            <div class="size-inputs">
                                <div class="size-input-group">
                                    <label>最大宽度</label>
                                    <input type="number" id="maxWidth" placeholder="不限制" min="0">
                                    <span>px</span>
                                </div>
                                <div class="size-input-group">
                                    <label>最大高度</label>
                                    <input type="number" id="maxHeight" placeholder="不限制" min="0">
                                    <span>px</span>
                                </div>
                            </div>
                            <label class="checkbox-label">
                                <input type="checkbox" id="keepAspect" checked>
                                <span>保持宽高比</span>
                            </label>
                        </div>

                        <div class="settings-section stats-section">
                            <h3>统计信息</h3>
                            <div class="stats-grid">
                                <div class="stat-item">
                                    <span class="stat-label">图片数量</span>
                                    <span class="stat-value" id="fileCount">0</span>
                                </div>
                                <div class="stat-item">
                                    <span class="stat-label">原始大小</span>
                                    <span class="stat-value" id="totalOriginal">0 KB</span>
                                </div>
                                <div class="stat-item">
                                    <span class="stat-label">压缩后</span>
                                    <span class="stat-value" id="totalCompressed">0 KB</span>
                                </div>
                                <div class="stat-item">
                                    <span class="stat-label">节省空间</span>
                                    <span class="stat-value stat-saved" id="totalSaved">0%</span>
                                </div>
                            </div>
                        </div>

                        <div class="compress-buttons">
                            <button class="btn btn-compress" id="compressBtn" disabled>
                                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <polyline points="16 16 12 12 8 16"/>
                                    <line x1="12" y1="12" x2="12" y2="21"/>
                                    <path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"/>
                                </svg>
                                开始压缩
                            </button>
                            <button class="btn btn-stop" id="stopCompressBtn" style="display: none;">
                                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <rect x="6" y="6" width="12" height="12"/>
                                </svg>
                                停止
                            </button>
                        </div>
                    </div>

                    <!-- GIF 模式设置 -->
                    <div id="gifSettings" style="display: none;">
                        <div class="settings-section">
                            <h3>帧率设置</h3>
                            <div class="frame-rate-control">
                                <div class="size-input-group">
                                    <label>帧延迟</label>
                                    <input type="number" id="frameDelay" value="100" min="10" max="5000">
                                    <span>毫秒</span>
                                </div>
                                <div class="fps-display">
                                    <span class="fps-value" id="fpsValue">10 FPS</span>
                                </div>
                            </div>
                            <input type="range" id="frameDelaySlider" min="20" max="500" value="100" class="delay-slider">
                            <div class="quality-labels">
                                <span>快速 (50fps)</span>
                                <span>慢速 (2fps)</span>
                            </div>
                        </div>

                        <div class="settings-section">
                            <h3>循环设置</h3>
                            <div class="loop-options">
                                <label class="checkbox-label">
                                    <input type="radio" name="loopType" id="loopInfinite" checked>
                                    <span>无限循环</span>
                                </label>
                                <label class="checkbox-label">
                                    <input type="radio" name="loopType" id="loopOnce">
                                    <span>播放一次</span>
                                </label>
                                <label class="checkbox-label">
                                    <input type="radio" name="loopType" id="loopCustom">
                                    <span>自定义次数</span>
                                    <input type="number" id="loopCount" value="3" min="1" max="100" style="width: 60px; margin-left: 8px;" disabled>
                                </label>
                            </div>
                        </div>

                        <div class="settings-section">
                            <h3>输出尺寸</h3>
                            <div class="size-inputs">
                                <div class="size-input-group">
                                    <label>最大宽度</label>
                                    <input type="number" id="gifMaxWidth" placeholder="不限制" min="0">
                                    <span>px</span>
                                </div>
                                <div class="size-input-group">
                                    <label>最大高度</label>
                                    <input type="number" id="gifMaxHeight" placeholder="不限制" min="0">
                                    <span>px</span>
                                </div>
                            </div>
                        </div>

                        <div class="settings-section">
                            <h3>文件名</h3>
                            <div class="size-input-group" style="flex: 1;">
                                <input type="text" id="gifOutputName" value="animation" placeholder="输出文件名" style="width: 100%;">
                                <span>.gif</span>
                            </div>
                        </div>

                        <div class="settings-section stats-section">
                            <h3>序列帧信息</h3>
                            <div class="stats-grid">
                                <div class="stat-item">
                                    <span class="stat-label">帧数</span>
                                    <span class="stat-value" id="gifFrameCount">0</span>
                                </div>
                                <div class="stat-item">
                                    <span class="stat-label">预计时长</span>
                                    <span class="stat-value" id="gifDuration">0 秒</span>
                                </div>
                            </div>
                        </div>

                        <button class="btn btn-compress btn-gif" id="createGifBtn" disabled>
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/>
                                <line x1="7" y1="2" x2="7" y2="22"/>
                                <line x1="17" y1="2" x2="17" y2="22"/>
                                <line x1="2" y1="12" x2="22" y2="12"/>
                            </svg>
                            生成 GIF
                        </button>
                    </div>
                </div>
            </div>

            <!-- 进度条 -->
            <div class="progress-bar" id="progressBar">
                <div class="progress-fill" id="progressFill"></div>
                <span class="progress-text" id="progressText">准备就绪</span>
            </div>
        </div>

        <!-- 预览模态框 -->
        <div class="modal" id="previewModal">
            <div class="modal-content">
                <button class="modal-close" id="closeModal">×</button>
                <div class="comparison-container" id="comparisonContainer">
                    <div class="comparison-after" id="comparisonAfter">
                        <img id="afterImg" alt="压缩后">
                    </div>
                    <div class="comparison-before" id="comparisonBefore">
                        <img id="beforeImg" alt="原图">
                    </div>
                    <div class="comparison-slider" id="comparisonSlider"></div>
                    <div class="comparison-labels">
                        <span class="label-before">原图 <span id="beforeSize"></span></span>
                        <span class="label-after">压缩后 <span id="afterSize"></span></span>
                    </div>
                </div>
                <div class="modal-info">
                    <span id="modalFileName"></span>
                    <span class="modal-savings" id="modalSavings"></span>
                </div>
            </div>
        </div>

        <!-- GIF 预览模态框 -->
        <div class="modal" id="gifPreviewModal">
            <div class="modal-content">
                <button class="modal-close" id="closeGifModal">×</button>
                <div class="gif-preview-container">
                    <img id="gifPreviewImg" alt="GIF 预览">
                </div>
                <div class="modal-info">
                    <span id="gifModalInfo"></span>
                    <span class="modal-savings" id="gifModalSize"></span>
                </div>
            </div>
        </div>
    `;

    bindEvents();
    setupDragAndDrop();
}

// 格式化文件大小
function formatSize(bytes) {
    if (bytes === 0) return '0 B';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
}

// 更新文件列表显示
function updateFileList() {
    const listEl = document.getElementById('fileList');
    const dropContent = document.getElementById('dropZoneContent');

    if (state.files.length === 0) {
        listEl.innerHTML = '';
        dropContent.style.display = 'flex';
        return;
    }

    dropContent.style.display = 'none';
    listEl.innerHTML = state.files.map((file, index) => {
        const statusClass = file.status || 'pending';
        const statusIcon = {
            'pending': '',
            'processing': '<div class="spinner"></div>',
            'success': '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#38ef7d" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>',
            'error': '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#ff6b6b" stroke-width="3"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>'
        }[statusClass];

        const savings = file.result ?
            `<span class="file-savings">-${file.result.compressionRatio?.toFixed(0) || 0}%</span>` : '';

        // 压缩中禁用删除按钮
        const removeDisabled = state.isProcessing ? 'disabled' : '';

        return `
            <div class="file-item ${statusClass}" data-index="${index}">
                <div class="file-preview">
                    ${file.preview ? `<img src="${file.preview}" alt="${file.name}">` : '<div class="no-preview"></div>'}
                </div>
                <div class="file-info">
                    <div class="file-name" title="${file.name}">${file.name}</div>
                    <div class="file-meta">
                        <span>${formatSize(file.size)}</span>
                        ${file.width ? `<span>${file.width}×${file.height}</span>` : ''}
                        <span class="file-format">${(file.format || '').toUpperCase()}</span>
                    </div>
                </div>
                <div class="file-status">
                    ${statusIcon}
                    ${savings}
                </div>
                <button class="file-remove" data-index="${index}" title="移除" ${removeDisabled}>
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                    </svg>
                </button>
            </div>
        `;
    }).join('');

    // 绑定移除按钮事件
    listEl.querySelectorAll('.file-remove').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            if (state.isProcessing) return; // 压缩中不允许删除
            const index = parseInt(btn.dataset.index);
            state.files.splice(index, 1);
            updateFileList();
            updateStats();
            updateCompressButton();
        });
    });

    // 绑定点击预览事件
    listEl.querySelectorAll('.file-item.success').forEach(item => {
        item.addEventListener('click', () => {
            const index = parseInt(item.dataset.index);
            const file = state.files[index];
            if (file.result) {
                showPreviewModal(file);
            }
        });
    });
}

// 更新统计信息
function updateStats() {
    const count = state.files.length;
    const totalOriginal = state.files.reduce((sum, f) => sum + (f.size || 0), 0);
    const totalCompressed = state.files.reduce((sum, f) => sum + (f.result?.newSize || f.size || 0), 0);
    const saved = totalOriginal > 0 ? ((totalOriginal - totalCompressed) / totalOriginal * 100) : 0;

    document.getElementById('fileCount').textContent = count;
    document.getElementById('totalOriginal').textContent = formatSize(totalOriginal);
    document.getElementById('totalCompressed').textContent = formatSize(totalCompressed);
    document.getElementById('totalSaved').textContent = saved > 0 ? `-${saved.toFixed(1)}%` : '0%';
}

// 更新压缩按钮状态
function updateCompressButton() {
    const btn = document.getElementById('compressBtn');
    const canCompress = state.files.length > 0 && state.outputDir && !state.isProcessing;
    btn.disabled = !canCompress;
    btn.innerHTML = state.isProcessing ?
        '<div class="spinner"></div> 压缩中...' :
        `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="16 16 12 12 8 16"/>
            <line x1="12" y1="12" x2="12" y2="21"/>
            <path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"/>
        </svg> 开始压缩`;
}

// 更新进度条
function updateProgress(current, total, text) {
    const progressBar = document.getElementById('progressBar');
    const progressFill = document.getElementById('progressFill');
    const progressText = document.getElementById('progressText');

    if (current === 0 && total === 0) {
        progressBar.classList.remove('active');
        progressFill.style.width = '0%';
        progressText.textContent = text || '准备就绪';
    } else {
        progressBar.classList.add('active');
        const percent = (current / total) * 100;
        progressFill.style.width = percent + '%';
        progressText.textContent = text || `处理中 ${current}/${total}`;
    }
}

// 显示预览模态框
function showPreviewModal(file) {
    const modal = document.getElementById('previewModal');
    const result = file.result;

    document.getElementById('beforeImg').src = result.originalBase64;
    document.getElementById('afterImg').src = result.compressedBase64;
    document.getElementById('beforeSize').textContent = formatSize(result.originalSize);
    document.getElementById('afterSize').textContent = formatSize(result.newSize);
    document.getElementById('modalFileName').textContent = file.name;
    document.getElementById('modalSavings').textContent = `-${result.compressionRatio.toFixed(1)}%`;

    modal.classList.add('active');

    // 初始化滑块
    initComparisonSlider();
}

// 初始化对比滑块
function initComparisonSlider() {
    const container = document.getElementById('comparisonContainer');
    const beforeEl = document.getElementById('comparisonBefore');
    const slider = document.getElementById('comparisonSlider');

    let isDragging = false;

    function updatePosition(x) {
        const rect = container.getBoundingClientRect();
        let percent = ((x - rect.left) / rect.width) * 100;
        percent = Math.max(0, Math.min(100, percent));
        beforeEl.style.width = percent + '%';
        slider.style.left = percent + '%';
    }

    const onMouseDown = (e) => {
        isDragging = true;
        updatePosition(e.clientX || e.touches[0].clientX);
    };

    const onMouseMove = (e) => {
        if (isDragging) {
            updatePosition(e.clientX || e.touches[0].clientX);
        }
    };

    const onMouseUp = () => {
        isDragging = false;
    };

    container.addEventListener('mousedown', onMouseDown);
    container.addEventListener('touchstart', onMouseDown);
    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('touchmove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
    document.addEventListener('touchend', onMouseUp);
}

// 绑定事件
function bindEvents() {
    // 模式切换
    document.querySelectorAll('.mode-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.mode-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            state.mode = btn.dataset.mode;
            updateModeUI();
        });
    });

    // 添加文件按钮
    document.getElementById('addFiles').addEventListener('click', async () => {
        const files = await SelectImages();
        if (files && files.length > 0) {
            await addFiles(files);
        }
    });

    // 选择输出目录
    document.getElementById('selectOutput').addEventListener('click', async () => {
        const dir = await SelectOutputDir();
        if (dir) {
            state.outputDir = dir;
            document.getElementById('outputPath').textContent = dir.split('/').slice(-2).join('/');
            document.getElementById('outputPath').title = dir;
            updateCompressButton();
            updateGifButton();
        }
    });

    // 清空列表
    document.getElementById('clearAll').addEventListener('click', () => {
        state.files = [];
        updateFileList();
        updateStats();
        updateCompressButton();
        updateGifButton();
    });

    // 质量滑块
    const qualitySlider = document.getElementById('qualitySlider');
    const qualityValue = document.getElementById('qualityValue');
    const qualityHint = document.getElementById('qualityHint');

    function updateQualityHint(val) {
        const format = state.options.outputFormat;

        // PNG 和 GIF 格式不支持质量调节
        if (format === 'png') {
            qualityHint.textContent = 'ℹ PNG 是无损格式，质量滑块无效。建议转为 JPEG/WebP';
            qualityHint.className = 'quality-hint info';
            return;
        }
        if (format === 'gif') {
            qualityHint.textContent = 'ℹ GIF 格式不支持质量调节';
            qualityHint.className = 'quality-hint info';
            return;
        }

        // 原格式时检查上传的图片格式
        if (format === 'original' && state.files.length > 0) {
            const formats = state.files.map(f => (f.format || '').toLowerCase());
            const allPng = formats.every(f => f === 'png');
            const allGif = formats.every(f => f === 'gif');
            const hasPngOrGif = formats.some(f => f === 'png' || f === 'gif');

            if (allPng) {
                qualityHint.textContent = 'ℹ PNG 是无损格式，质量滑块无效。建议转为 JPEG/WebP';
                qualityHint.className = 'quality-hint info';
                return;
            }
            if (allGif) {
                qualityHint.textContent = 'ℹ GIF 格式不支持质量调节';
                qualityHint.className = 'quality-hint info';
                return;
            }
            if (hasPngOrGif) {
                qualityHint.textContent = 'ℹ 列表中有 PNG/GIF，这些图片质量滑块无效';
                qualityHint.className = 'quality-hint info';
                return;
            }
        }

        if (val >= 95) {
            qualityHint.textContent = '✓ 接近无损，画质最佳';
            qualityHint.className = 'quality-hint lossless';
        } else if (val >= 80) {
            qualityHint.textContent = '推荐：画质与体积平衡';
            qualityHint.className = 'quality-hint good';
        } else if (val >= 50) {
            qualityHint.textContent = '适中：有轻微压缩痕迹';
            qualityHint.className = 'quality-hint medium';
        } else {
            qualityHint.textContent = '⚠ 低质量：图片会明显模糊';
            qualityHint.className = 'quality-hint low';
        }
    }

    // 输出格式选择（放在 updateQualityHint 定义之后）
    document.querySelectorAll('.format-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.format-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            state.options.outputFormat = btn.dataset.format;
            // 更新质量提示（PNG/GIF 不支持质量调节）
            updateQualityHint(state.options.quality);
        });
    });

    qualitySlider.addEventListener('input', () => {
        const val = parseInt(qualitySlider.value);
        state.options.quality = val;
        qualityValue.textContent = val + '%';
        updateQualityHint(val);
    });

    // 初始化提示
    updateQualityHint(80);

    // 尺寸输入
    document.getElementById('maxWidth').addEventListener('change', (e) => {
        state.options.maxWidth = parseInt(e.target.value) || 0;
    });
    document.getElementById('maxHeight').addEventListener('change', (e) => {
        state.options.maxHeight = parseInt(e.target.value) || 0;
    });

    // 保持宽高比
    document.getElementById('keepAspect').addEventListener('change', (e) => {
        state.options.keepAspect = e.target.checked;
    });

    // 开始压缩
    document.getElementById('compressBtn').addEventListener('click', startCompression);

    // 停止压缩
    document.getElementById('stopCompressBtn').addEventListener('click', stopCompression);

    // GIF 设置事件
    bindGifEvents();

    // 关闭模态框
    document.getElementById('closeModal').addEventListener('click', () => {
        document.getElementById('previewModal').classList.remove('active');
    });
    document.getElementById('previewModal').addEventListener('click', (e) => {
        if (e.target.id === 'previewModal') {
            document.getElementById('previewModal').classList.remove('active');
        }
    });

    // GIF 预览模态框
    document.getElementById('closeGifModal').addEventListener('click', () => {
        document.getElementById('gifPreviewModal').classList.remove('active');
    });
    document.getElementById('gifPreviewModal').addEventListener('click', (e) => {
        if (e.target.id === 'gifPreviewModal') {
            document.getElementById('gifPreviewModal').classList.remove('active');
        }
    });
}

// GIF 设置事件绑定
function bindGifEvents() {
    // 帧延迟输入
    const frameDelayInput = document.getElementById('frameDelay');
    const frameDelaySlider = document.getElementById('frameDelaySlider');
    const fpsValue = document.getElementById('fpsValue');

    const updateFpsDisplay = (delay) => {
        const fps = Math.round(1000 / delay);
        fpsValue.textContent = fps + ' FPS';
        state.gifOptions.frameDelay = delay;
        updateGifStats();
    };

    frameDelayInput.addEventListener('change', (e) => {
        const val = Math.max(10, Math.min(5000, parseInt(e.target.value) || 100));
        frameDelayInput.value = val;
        frameDelaySlider.value = Math.min(500, Math.max(20, val));
        updateFpsDisplay(val);
    });

    frameDelaySlider.addEventListener('input', () => {
        const val = parseInt(frameDelaySlider.value);
        frameDelayInput.value = val;
        updateFpsDisplay(val);
    });

    // 循环设置
    document.getElementById('loopInfinite').addEventListener('change', () => {
        state.gifOptions.loopCount = 0;
        document.getElementById('loopCount').disabled = true;
    });
    document.getElementById('loopOnce').addEventListener('change', () => {
        state.gifOptions.loopCount = 1;
        document.getElementById('loopCount').disabled = true;
    });
    document.getElementById('loopCustom').addEventListener('change', () => {
        document.getElementById('loopCount').disabled = false;
        state.gifOptions.loopCount = parseInt(document.getElementById('loopCount').value) || 3;
    });
    document.getElementById('loopCount').addEventListener('change', (e) => {
        state.gifOptions.loopCount = parseInt(e.target.value) || 3;
    });

    // GIF 尺寸设置
    document.getElementById('gifMaxWidth').addEventListener('change', (e) => {
        state.gifOptions.maxWidth = parseInt(e.target.value) || 0;
    });
    document.getElementById('gifMaxHeight').addEventListener('change', (e) => {
        state.gifOptions.maxHeight = parseInt(e.target.value) || 0;
    });

    // 输出文件名
    document.getElementById('gifOutputName').addEventListener('change', (e) => {
        state.gifOptions.outputName = e.target.value || 'animation';
    });

    // 生成 GIF 按钮
    document.getElementById('createGifBtn').addEventListener('click', createGif);
}

// 更新模式 UI
function updateModeUI() {
    const compressSettings = document.getElementById('compressSettings');
    const gifSettings = document.getElementById('gifSettings');
    const dropZoneTitle = document.getElementById('dropZoneTitle');
    const dropZoneSubtitle = document.getElementById('dropZoneSubtitle');

    if (state.mode === 'compress') {
        compressSettings.style.display = 'block';
        gifSettings.style.display = 'none';
        dropZoneTitle.textContent = '拖放图片到这里';
        dropZoneSubtitle.textContent = '或点击上方按钮选择文件';
    } else if (state.mode === 'gif') {
        compressSettings.style.display = 'none';
        gifSettings.style.display = 'block';
        dropZoneTitle.textContent = '拖放序列帧到这里';
        dropZoneSubtitle.textContent = '按文件名顺序自动排序（支持数字排序）';
    }

    updateGifStats();
    updateGifButton();
}

// 更新 GIF 统计信息
function updateGifStats() {
    const frameCount = state.files.length;
    const duration = (frameCount * state.gifOptions.frameDelay / 1000).toFixed(1);

    const gifFrameCount = document.getElementById('gifFrameCount');
    const gifDuration = document.getElementById('gifDuration');

    if (gifFrameCount) gifFrameCount.textContent = frameCount;
    if (gifDuration) gifDuration.textContent = duration + ' 秒';
}

// 更新 GIF 按钮状态
function updateGifButton() {
    const btn = document.getElementById('createGifBtn');
    if (!btn) return;

    const canCreate = state.files.length >= 2 && state.outputDir && !state.isProcessing;
    btn.disabled = !canCreate;
    btn.innerHTML = state.isProcessing ?
        '<div class="spinner"></div> 生成中...' :
        `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/>
            <line x1="7" y1="2" x2="7" y2="22"/>
            <line x1="17" y1="2" x2="17" y2="22"/>
            <line x1="2" y1="12" x2="22" y2="12"/>
        </svg> 生成 GIF`;
}

// 创建 GIF
async function createGif() {
    if (state.isProcessing || state.files.length < 2 || !state.outputDir) return;

    state.isProcessing = true;
    updateGifButton();
    updateProgress(1, 2, '正在生成 GIF...');

    try {
        const paths = state.files.map(f => f.path);
        const result = await CreateGifFromSequence(paths, {
            frameDelay: state.gifOptions.frameDelay,
            loopCount: state.gifOptions.loopCount,
            maxWidth: state.gifOptions.maxWidth,
            maxHeight: state.gifOptions.maxHeight,
            outputDir: state.outputDir,
            outputName: state.gifOptions.outputName
        });

        if (result.success) {
            updateProgress(2, 2, 'GIF 创建成功！');
            showGifPreviewModal(result);
        } else {
            updateProgress(0, 0);
            alert('GIF 创建失败: ' + result.message);
        }
    } catch (err) {
        console.error(err);
        updateProgress(0, 0);
        alert('GIF 创建出错: ' + err.message);
    }

    state.isProcessing = false;
    updateGifButton();
    setTimeout(() => updateProgress(0, 0), 2000);
}

// 显示 GIF 预览模态框
function showGifPreviewModal(result) {
    const modal = document.getElementById('gifPreviewModal');
    document.getElementById('gifPreviewImg').src = result.preview;
    document.getElementById('gifModalInfo').textContent =
        `${result.frameCount} 帧 | ${result.width}×${result.height}`;
    document.getElementById('gifModalSize').textContent = formatSize(result.fileSize);
    modal.classList.add('active');
}

// 设置拖放
function setupDragAndDrop() {
    const dropZone = document.getElementById('dropZone');

    // 监听 Wails 的文件拖放事件
    EventsOn('wails:file-drop', async (x, y, paths) => {
        if (paths && paths.length > 0) {
            // 过滤图片文件
            const imageExts = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.tiff', '.tif', '.bmp'];
            const imagePaths = paths.filter(p => {
                const ext = p.toLowerCase().substring(p.lastIndexOf('.'));
                return imageExts.includes(ext);
            });
            if (imagePaths.length > 0) {
                await addFiles(imagePaths);
            }
        }
    });

    // 拖放视觉反馈
    dropZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropZone.classList.add('drag-over');
    });

    dropZone.addEventListener('dragleave', () => {
        dropZone.classList.remove('drag-over');
    });

    dropZone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropZone.classList.remove('drag-over');
    });
}

// 添加文件
async function addFiles(paths) {
    for (const path of paths) {
        // 检查是否已存在
        if (state.files.find(f => f.path === path)) continue;

        const info = await GetImageInfo(path);
        state.files.push({
            path: info.path,
            name: info.name,
            size: info.size,
            width: info.width,
            height: info.height,
            format: info.format,
            preview: info.preview,
            status: 'pending',
            result: null
        });
    }
    updateFileList();
    updateStats();
    updateCompressButton();
    updateGifStats();
    updateGifButton();
    // 触发质量提示更新（检测上传的图片格式）
    const qualitySlider = document.getElementById('qualitySlider');
    if (qualitySlider) qualitySlider.dispatchEvent(new Event('input'));
}

// 开始压缩
// 停止压缩
function stopCompression() {
    state.stopRequested = true;
    updateProgress(0, 0, '正在停止...');
}

// 显示/隐藏停止按钮
function updateStopButtons(show) {
    const stopBtn = document.getElementById('stopCompressBtn');
    if (stopBtn) stopBtn.style.display = show ? 'flex' : 'none';
}

async function startCompression() {
    if (state.isProcessing || state.files.length === 0 || !state.outputDir) return;

    state.isProcessing = true;
    state.stopRequested = false;
    updateCompressButton();
    updateStopButtons(true);
    updateFileList();  // 刷新列表，禁用删除按钮

    const total = state.files.length;
    let processed = 0;

    for (let i = 0; i < state.files.length; i++) {
        // 检查是否请求停止
        if (state.stopRequested) {
            break;
        }

        const file = state.files[i];
        if (file.status === 'success') {
            processed++;
            continue;
        }

        file.status = 'processing';
        updateFileList();
        updateProgress(processed, total, `正在压缩: ${file.name}`);

        try {
            const result = await CompressImage(file.path, {
                quality: state.options.quality,
                maxWidth: state.options.maxWidth,
                maxHeight: state.options.maxHeight,
                outputFormat: state.options.outputFormat,
                outputDir: state.outputDir,
                keepAspect: state.options.keepAspect
            });

            if (result.success) {
                file.status = 'success';
                file.result = result;
            } else {
                file.status = 'error';
                console.error(result.message);
            }
        } catch (err) {
            file.status = 'error';
            console.error(err);
        }

        processed++;
        updateFileList();
        updateStats();
    }

    // 如果被停止，将 processing 状态的文件重置为 pending
    if (state.stopRequested) {
        state.files.forEach(f => {
            if (f.status === 'processing') {
                f.status = 'pending';
            }
        });
        updateProgress(0, 0, '已停止');
    } else {
        updateProgress(0, 0, '压缩完成!');
    }

    state.isProcessing = false;
    state.stopRequested = false;
    updateCompressButton();
    updateStopButtons(false);
    updateFileList();  // 刷新列表，启用删除按钮
}

// 初始化
initUI();
