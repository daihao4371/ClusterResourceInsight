class ClusterManager {
    constructor() {
        this.apiBase = '/api/v1';
        this.currentCluster = null;
        this.clusters = [];
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadClusters();
    }

    bindEvents() {
        // 主要按钮事件
        document.getElementById('add-cluster-btn').addEventListener('click', () => {
            this.showAddClusterModal();
        });

        document.getElementById('test-all-btn').addEventListener('click', () => {
            this.testAllClusters();
        });

        document.getElementById('retry-btn').addEventListener('click', () => {
            this.loadClusters();
        });

        // 模态框事件
        document.getElementById('modal-close').addEventListener('click', () => {
            this.hideModal();
        });

        document.getElementById('cancel-btn').addEventListener('click', () => {
            this.hideModal();
        });

        document.getElementById('cluster-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveCluster();
        });

        // 认证方式切换事件
        document.getElementById('auth-type').addEventListener('change', (e) => {
            this.showAuthConfig(e.target.value);
        });

        // 测试连接按钮
        document.getElementById('test-connection-btn').addEventListener('click', () => {
            this.testConnection();
        });

        // 删除确认模态框事件
        document.getElementById('delete-modal-close').addEventListener('click', () => {
            this.hideDeleteModal();
        });

        document.getElementById('cancel-delete-btn').addEventListener('click', () => {
            this.hideDeleteModal();
        });

        document.getElementById('confirm-delete-btn').addEventListener('click', () => {
            this.confirmDelete();
        });

        // 模态框外部点击关闭
        window.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal')) {
                e.target.style.display = 'none';
            }
        });
    }

    async loadClusters() {
        this.showLoading();
        
        try {
            const response = await axios.get(`${this.apiBase}/clusters`);
            this.clusters = response.data.data || [];
            
            this.updateSummary();
            this.renderClusters();
            this.showClusters();
            
        } catch (error) {
            console.error('Error loading clusters:', error);
            this.showError('加载集群列表失败: ' + (error.response?.data?.error || error.message));
        }
    }

    updateSummary() {
        const total = this.clusters.length;
        const online = this.clusters.filter(c => c.status === 'online').length;
        const offline = total - online;

        document.getElementById('total-clusters').textContent = total;
        document.getElementById('online-clusters').textContent = online;
        document.getElementById('offline-clusters').textContent = offline;
    }

    renderClusters() {
        const container = document.getElementById('clusters-grid');
        container.innerHTML = '';

        if (this.clusters.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">🖥️</div>
                    <h3>暂无集群</h3>
                    <p>点击"添加集群"按钮来添加您的第一个K8s集群</p>
                    <button class="btn btn-primary" onclick="document.getElementById('add-cluster-btn').click()">
                        添加集群
                    </button>
                </div>
            `;
            return;
        }

        this.clusters.forEach(cluster => {
            const card = this.createClusterCard(cluster);
            container.appendChild(card);
        });
    }

    createClusterCard(cluster) {
        const card = document.createElement('div');
        card.className = 'cluster-card';
        card.dataset.clusterId = cluster.id;

        const statusClass = cluster.status === 'online' ? 'online' : 
                           cluster.status === 'offline' ? 'offline' : 'unknown';
        
        const lastCollectText = cluster.last_collect_at ? 
            `最后采集: ${new Date(cluster.last_collect_at).toLocaleString('zh-CN')}` : 
            '暂未采集';

        card.innerHTML = `
            <div class="cluster-header">
                <div class="cluster-info">
                    <h3 class="cluster-name">${cluster.cluster_name}</h3>
                    <p class="cluster-alias">${cluster.cluster_alias || '无别名'}</p>
                </div>
                <div class="cluster-status ${statusClass}">
                    <span class="status-dot"></span>
                    <span class="status-text">${this.getStatusText(cluster.status)}</span>
                </div>
            </div>
            
            <div class="cluster-details">
                <div class="detail-item">
                    <span class="label">API Server:</span>
                    <span class="value">${cluster.api_server}</span>
                </div>
                <div class="detail-item">
                    <span class="label">认证方式:</span>
                    <span class="value">${this.getAuthTypeText(cluster.auth_type)}</span>
                </div>
                <div class="detail-item">
                    <span class="label">采集间隔:</span>
                    <span class="value">${cluster.collect_interval} 分钟</span>
                </div>
                <div class="detail-item">
                    <span class="label">${lastCollectText}</span>
                </div>
                ${cluster.tags && cluster.tags !== '[]' ? `
                <div class="cluster-tags">
                    ${this.parseTags(cluster.tags).map(tag => `<span class="tag">${tag}</span>`).join('')}
                </div>
                ` : ''}
            </div>
            
            <div class="cluster-actions">
                <button class="btn btn-sm btn-outline" onclick="clusterManager.testCluster(${cluster.id})">
                    <span class="icon">🔍</span>
                    测试
                </button>
                <button class="btn btn-sm btn-outline" onclick="clusterManager.editCluster(${cluster.id})">
                    <span class="icon">✏️</span>
                    编辑
                </button>
                <button class="btn btn-sm btn-danger" onclick="clusterManager.deleteCluster(${cluster.id})">
                    <span class="icon">🗑️</span>
                    删除
                </button>
            </div>
        `;

        return card;
    }

    getStatusText(status) {
        switch (status) {
            case 'online': return '在线';
            case 'offline': return '离线';
            case 'unknown': return '未知';
            default: return status || '未知';
        }
    }

    getAuthTypeText(authType) {
        switch (authType) {
            case 'token': return 'Bearer Token';
            case 'cert': return '客户端证书';
            case 'kubeconfig': return 'Kubeconfig';
            default: return authType || '未知';
        }
    }

    parseTags(tagsString) {
        try {
            return JSON.parse(tagsString) || [];
        } catch {
            return [];
        }
    }

    // 显示添加集群模态框
    showAddClusterModal() {
        document.getElementById('modal-title').textContent = '添加集群';
        document.getElementById('cluster-form').reset();
        this.currentCluster = null;
        this.showAuthConfig('');
        document.getElementById('cluster-modal').style.display = 'flex';
    }

    // 编辑集群
    async editCluster(clusterId) {
        try {
            const response = await axios.get(`${this.apiBase}/clusters/${clusterId}`);
            const cluster = response.data;
            
            this.currentCluster = cluster;
            document.getElementById('modal-title').textContent = '编辑集群';
            
            // 填充表单
            document.getElementById('cluster-name').value = cluster.cluster_name;
            document.getElementById('cluster-alias').value = cluster.cluster_alias || '';
            document.getElementById('api-server').value = cluster.api_server;
            document.getElementById('auth-type').value = cluster.auth_type;
            document.getElementById('collect-interval').value = cluster.collect_interval;
            
            // 处理标签
            const tags = this.parseTags(cluster.tags);
            document.getElementById('tags').value = tags.join(', ');
            
            this.showAuthConfig(cluster.auth_type);
            
            document.getElementById('cluster-modal').style.display = 'flex';
            
        } catch (error) {
            console.error('Error loading cluster details:', error);
            alert('加载集群详情失败: ' + (error.response?.data?.error || error.message));
        }
    }

    // 删除集群
    deleteCluster(clusterId) {
        const cluster = this.clusters.find(c => c.id === clusterId);
        if (!cluster) return;
        
        document.getElementById('delete-cluster-name').textContent = cluster.cluster_name;
        this.currentCluster = cluster;
        document.getElementById('delete-modal').style.display = 'flex';
    }

    // 确认删除
    async confirmDelete() {
        if (!this.currentCluster) return;
        
        try {
            await axios.delete(`${this.apiBase}/clusters/${this.currentCluster.id}`);
            this.hideDeleteModal();
            await this.loadClusters();
            this.showToast('集群删除成功', 'success');
            
        } catch (error) {
            console.error('Error deleting cluster:', error);
            alert('删除集群失败: ' + (error.response?.data?.error || error.message));
        }
    }

    // 测试单个集群
    async testCluster(clusterId) {
        try {
            const card = document.querySelector(`[data-cluster-id="${clusterId}"]`);
            if (card) {
                card.classList.add('testing');
            }
            
            const response = await axios.post(`${this.apiBase}/clusters/${clusterId}/test`);
            const result = response.data.data; // 修复：从 data 字段中获取结果
            
            if (card) {
                card.classList.remove('testing');
                
                // 更新状态显示
                const statusElement = card.querySelector('.cluster-status');
                statusElement.className = `cluster-status ${result.status}`;
                statusElement.querySelector('.status-text').textContent = this.getStatusText(result.status);
            }
            
            this.showToast(`集群测试完成: ${result.message}`, result.success ? 'success' : 'error');
            
        } catch (error) {
            console.error('Error testing cluster:', error);
            const card = document.querySelector(`[data-cluster-id="${clusterId}"]`);
            if (card) {
                card.classList.remove('testing');
            }
            
            let errorMessage = '测试失败';
            
            // 改进错误消息解析
            if (error.response?.data?.error) {
                errorMessage = error.response.data.error;
            } else if (error.response?.data?.message) {
                errorMessage = error.response.data.message;
            } else if (error.message) {
                errorMessage = error.message;
            }
            
            this.showToast('测试集群失败: ' + errorMessage, 'error');
        }
    }

    // 批量测试所有集群
    async testAllClusters() {
        try {
            document.getElementById('test-all-btn').disabled = true;
            document.getElementById('test-all-btn').textContent = '测试中...';
            
            const response = await axios.post(`${this.apiBase}/clusters/batch-test`);
            const results = response.data;
            
            // 更新集群状态显示
            Object.entries(results).forEach(([clusterId, result]) => {
                const card = document.querySelector(`[data-cluster-id="${clusterId}"]`);
                if (card) {
                    const statusElement = card.querySelector('.cluster-status');
                    statusElement.className = `cluster-status ${result.status}`;
                    statusElement.querySelector('.status-text').textContent = this.getStatusText(result.status);
                }
            });
            
            this.showToast('批量测试完成', 'success');
            await this.loadClusters(); // 重新加载以获取最新状态
            
        } catch (error) {
            console.error('Error batch testing clusters:', error);
            alert('批量测试失败: ' + (error.response?.data?.error || error.message));
        } finally {
            document.getElementById('test-all-btn').disabled = false;
            document.getElementById('test-all-btn').innerHTML = '<span class="icon">🔄</span>批量测试';
        }
    }

    // 根据认证方式显示对应的配置区域
    showAuthConfig(authType) {
        // 隐藏所有认证配置区域
        document.querySelectorAll('.auth-config').forEach(el => {
            el.style.display = 'none';
        });
        
        // 显示对应的认证配置区域
        if (authType) {
            const configElement = document.getElementById(`${authType}-config`);
            if (configElement) {
                configElement.style.display = 'block';
            }
        }
    }

    // 测试连接（创建前测试）
    async testConnection() {
        const formData = this.getFormData();
        if (!this.validateForm(formData)) return;
        
        try {
            document.getElementById('test-connection-btn').disabled = true;
            document.getElementById('test-connection-btn').textContent = '测试中...';
            
            const response = await axios.post(`${this.apiBase}/clusters/test`, formData);
            const result = response.data.data; // 修复：从 data 字段中获取结果
            
            this.showToast(
                `连接测试${result.success ? '成功' : '失败'}: ${result.message}`, 
                result.success ? 'success' : 'error'
            );
            
        } catch (error) {
            console.error('Error testing connection:', error);
            let errorMessage = '连接失败';
            
            // 改进错误消息解析
            if (error.response?.data?.error) {
                errorMessage = error.response.data.error;
            } else if (error.response?.data?.message) {
                errorMessage = error.response.data.message;
            } else if (error.message) {
                errorMessage = error.message;
            }
            
            this.showToast('连接测试失败: ' + errorMessage, 'error');
        } finally {
            document.getElementById('test-connection-btn').disabled = false;
            document.getElementById('test-connection-btn').textContent = '测试连接';
        }
    }

    // 保存集群
    async saveCluster() {
        const formData = this.getFormData();
        if (!this.validateForm(formData)) return;
        
        try {
            document.getElementById('save-btn').disabled = true;
            document.getElementById('save-btn').textContent = '保存中...';
            
            if (this.currentCluster) {
                // 更新集群
                await axios.put(`${this.apiBase}/clusters/${this.currentCluster.id}`, formData);
                this.showToast('集群更新成功', 'success');
            } else {
                // 创建新集群
                await axios.post(`${this.apiBase}/clusters`, formData);
                this.showToast('集群创建成功', 'success');
            }
            
            this.hideModal();
            await this.loadClusters();
            
        } catch (error) {
            console.error('Error saving cluster:', error);
            this.showToast('保存集群失败: ' + (error.response?.data?.error || error.message), 'error');
        } finally {
            document.getElementById('save-btn').disabled = false;
            document.getElementById('save-btn').textContent = '保存';
        }
    }

    // 获取表单数据
    getFormData() {
        const authType = document.getElementById('auth-type').value;
        const authConfig = {};
        
        // 根据认证方式获取对应的配置
        switch (authType) {
            case 'token':
                authConfig.bearer_token = document.getElementById('bearer-token').value;
                break;
            case 'cert':
                authConfig.client_cert = document.getElementById('client-cert').value;
                authConfig.client_key = document.getElementById('client-key').value;
                authConfig.ca_cert = document.getElementById('ca-cert').value;
                break;
            case 'kubeconfig':
                authConfig.kubeconfig = document.getElementById('kubeconfig').value;
                break;
        }
        
        // 处理标签
        const tagsString = document.getElementById('tags').value.trim();
        const tags = tagsString ? tagsString.split(',').map(t => t.trim()).filter(t => t) : [];
        
        return {
            cluster_name: document.getElementById('cluster-name').value.trim(),
            cluster_alias: document.getElementById('cluster-alias').value.trim(),
            api_server: document.getElementById('api-server').value.trim(),
            auth_type: authType,
            auth_config: authConfig,
            tags: tags,
            collect_interval: parseInt(document.getElementById('collect-interval').value) || 30
        };
    }

    // 验证表单
    validateForm(formData) {
        if (!formData.cluster_name.trim()) {
            this.showToast('请输入集群名称', 'error');
            return false;
        }
        
        if (!formData.api_server.trim()) {
            this.showToast('请输入API Server地址', 'error');
            return false;
        }
        
        if (!formData.auth_type) {
            this.showToast('请选择认证方式', 'error');
            return false;
        }
        
        // 验证认证配置
        switch (formData.auth_type) {
            case 'token':
                if (!formData.auth_config.bearer_token.trim()) {
                    this.showToast('请输入Bearer Token', 'error');
                    return false;
                }
                break;
            case 'cert':
                if (!formData.auth_config.client_cert.trim() || !formData.auth_config.client_key.trim()) {
                    this.showToast('请输入客户端证书和私钥', 'error');
                    return false;
                }
                break;
            case 'kubeconfig':
                if (!formData.auth_config.kubeconfig.trim()) {
                    this.showToast('请输入Kubeconfig内容', 'error');
                    return false;
                }
                break;
        }
        
        return true;
    }

    // 显示/隐藏模态框
    hideModal() {
        document.getElementById('cluster-modal').style.display = 'none';
    }

    hideDeleteModal() {
        document.getElementById('delete-modal').style.display = 'none';
        this.currentCluster = null;
    }

    // 显示状态函数
    showLoading() {
        document.getElementById('loading').style.display = 'block';
        document.getElementById('clusters-container').style.display = 'none';
        document.getElementById('error-container').style.display = 'none';
    }

    showClusters() {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('clusters-container').style.display = 'block';
        document.getElementById('error-container').style.display = 'none';
    }

    showError(message) {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('clusters-container').style.display = 'none';
        document.getElementById('error-container').style.display = 'block';
        document.getElementById('error-message').textContent = message;
    }

    // 显示提示消息
    showToast(message, type = 'info') {
        // 创建 toast 元素
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        
        // 添加到页面
        document.body.appendChild(toast);
        
        // 显示动画
        setTimeout(() => toast.classList.add('show'), 100);
        
        // 自动隐藏
        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => document.body.removeChild(toast), 300);
        }, 3000);
    }
}

// 全局变量以便 HTML 中的 onclick 事件可以访问
let clusterManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    clusterManager = new ClusterManager();
});