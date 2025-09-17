class ResourceMonitor {
    constructor() {
        this.apiBase = '/api/v1';
        this.currentData = null;
        this.clusters = [];
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadClusters();
        this.loadData();
    }

    bindEvents() {
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadData();
        });

        document.getElementById('export-btn').addEventListener('click', () => {
            this.exportData();
        });

        document.getElementById('retry-btn').addEventListener('click', () => {
            this.loadData();
        });

        // 集群筛选事件
        document.getElementById('cluster-filter').addEventListener('change', (e) => {
            this.filterByCluster(e.target.value);
        });
    }

    // 加载集群列表用于筛选器
    async loadClusters() {
        try {
            const response = await axios.get(`${this.apiBase}/clusters`);
            this.clusters = response.data.data || [];
            this.updateClusterFilter();
        } catch (error) {
            console.error('Error loading clusters:', error);
        }
    }

    // 更新集群筛选下拉框
    updateClusterFilter() {
        const select = document.getElementById('cluster-filter');
        
        // 清除现有选项（保留"所有集群"）
        while (select.children.length > 1) {
            select.removeChild(select.lastChild);
        }

        // 添加集群选项
        this.clusters.forEach(cluster => {
            const option = document.createElement('option');
            option.value = cluster.cluster_name;
            option.textContent = cluster.cluster_alias || cluster.cluster_name;
            select.appendChild(option);
        });
    }

    async loadData() {
        this.showLoading();
        
        try {
            const response = await axios.get(`${this.apiBase}/analysis`);
            const data = response.data;
            
            this.currentData = data;
            this.updateSummary(data);
            this.renderTable(data.top50_problems || []);
            this.showTable();
            
        } catch (error) {
            console.error('Error loading data:', error);
            this.showError('加载数据失败: ' + (error.response?.data?.error || error.message));
        }
    }

    updateSummary(data) {
        document.getElementById('analyzed-clusters').textContent = data.clusters_analyzed || 0;
        document.getElementById('total-pods').textContent = data.total_pods || 0;
        document.getElementById('unreasonable-pods').textContent = data.unreasonable_pods || 0;
        document.getElementById('update-time').textContent = new Date(data.generated_at).toLocaleString('zh-CN');
    }

    renderTable(pods) {
        const tbody = document.getElementById('table-body');
        tbody.innerHTML = '';

        if (pods.length === 0) {
            tbody.innerHTML = '<tr><td colspan="16" style="text-align: center; padding: 20px;">暂无数据或所有Pod配置合理</td></tr>';
            return;
        }

        // 检查是否有 Metrics 数据
        const hasMetricsData = pods.some(pod => pod.memory_usage > 0 || pod.cpu_usage > 0);
        
        if (!hasMetricsData) {
            // 添加提示信息
            const warningRow = document.createElement('tr');
            warningRow.innerHTML = `
                <td colspan="16" style="text-align: center; padding: 15px; background-color: #fff3cd; color: #856404; border: 1px solid #ffeeba;">
                    <strong>⚠️ 提示:</strong> 当前集群未安装或未配置 Metrics Server，无法获取 Pod 实际资源使用量。
                    系统正在基于资源配置进行分析。建议安装 Metrics Server 以获得完整的资源利用率分析。
                </td>
            `;
            tbody.appendChild(warningRow);
        }

        pods.forEach(pod => {
            const row = this.createTableRow(pod);
            tbody.appendChild(row);
        });
    }

    createTableRow(pod) {
        const row = document.createElement('tr');
        if (pod.status === '不合理') {
            row.classList.add('unreasonable');
        }

        row.innerHTML = `
            <td class="cluster-name">${pod.cluster_name || '-'}</td>
            <td class="pod-name">${pod.pod_name}</td>
            <td class="namespace">${pod.namespace}</td>
            <td>${pod.node_name || '-'}</td>
            <td class="memory-usage">${this.formatBytes(pod.memory_usage)}</td>
            <td class="memory-usage">${this.formatBytes(pod.memory_request)}</td>
            <td class="percentage ${this.getPercentageClass(pod.memory_req_pct)}">${this.formatPercentage(pod.memory_req_pct)}</td>
            <td class="memory-usage">${this.formatBytes(pod.memory_limit)}</td>
            <td class="percentage ${this.getPercentageClass(pod.memory_limit_pct)}">${this.formatPercentage(pod.memory_limit_pct)}</td>
            <td class="cpu-usage">${this.formatMillicores(pod.cpu_usage)}</td>
            <td class="cpu-usage">${this.formatMillicores(pod.cpu_request)}</td>
            <td class="percentage ${this.getPercentageClass(pod.cpu_req_pct)}">${this.formatPercentage(pod.cpu_req_pct)}</td>
            <td class="cpu-usage">${this.formatMillicores(pod.cpu_limit)}</td>
            <td class="percentage ${this.getPercentageClass(pod.cpu_limit_pct)}">${this.formatPercentage(pod.cpu_limit_pct)}</td>
            <td class="${pod.status === '不合理' ? 'status-unreasonable' : 'status-reasonable'}">${pod.status}</td>
            <td class="issues">${(pod.issues || []).join(', ')}</td>
        `;

        return row;
    }

    // 根据集群筛选数据
    filterByCluster(clusterName) {
        if (!this.currentData || !this.currentData.top50_problems) {
            return;
        }

        let filteredPods = this.currentData.top50_problems;
        
        if (clusterName) {
            filteredPods = this.currentData.top50_problems.filter(pod => 
                pod.cluster_name === clusterName
            );
        }

        this.renderTable(filteredPods);
    }

    formatBytes(bytes) {
        if (!bytes || bytes === 0) return '<span style="color: #999;" title="Metrics数据不可用">-</span>';
        
        const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
        let size = bytes;
        let unitIndex = 0;
        
        while (size >= 1024 && unitIndex < units.length - 1) {
            size /= 1024;
            unitIndex++;
        }
        
        return `${size.toFixed(2)} ${units[unitIndex]}`;
    }

    formatMillicores(millicores) {
        if (!millicores || millicores === 0) return '<span style="color: #999;" title="Metrics数据不可用">-</span>';
        
        if (millicores >= 1000) {
            return `${(millicores / 1000).toFixed(2)}`;
        }
        return `${millicores}m`;
    }

    formatPercentage(percentage) {
        if (percentage === undefined || percentage === null || percentage === 0) {
            return '<span style="color: #999;" title="Metrics数据不可用或无基准值">-</span>';
        }
        return `${percentage.toFixed(3)}%`;
    }

    getPercentageClass(percentage) {
        if (percentage === undefined || percentage === null) return '';
        
        if (percentage < 20) return 'low';
        if (percentage < 50) return 'medium';
        return 'high';
    }

    showLoading() {
        document.getElementById('loading').style.display = 'block';
        document.getElementById('table-container').style.display = 'none';
        document.getElementById('error-container').style.display = 'none';
    }

    showTable() {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('table-container').style.display = 'block';
        document.getElementById('error-container').style.display = 'none';
    }

    showError(message) {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('table-container').style.display = 'none';
        document.getElementById('error-container').style.display = 'block';
        document.getElementById('error-message').textContent = message;
    }

    async exportData() {
        try {
            const response = await axios.get(`${this.apiBase}/analysis`);
            const data = response.data;
            
            const csv = this.convertToCSV(data.top50_problems || []);
            this.downloadCSV(csv, 'k8s-resource-analysis.csv');
            
        } catch (error) {
            console.error('Export error:', error);
            alert('导出失败: ' + (error.response?.data?.error || error.message));
        }
    }

    convertToCSV(pods) {
        const headers = [
            '集群名称', 'Pod名称', '命名空间', '节点', 
            '内存使用(bytes)', '内存请求(bytes)', '内存请求%', '内存限制(bytes)', '内存限制%',
            'CPU使用(millicores)', 'CPU请求(millicores)', 'CPU请求%', 'CPU限制(millicores)', 'CPU限制%',
            '状态', '问题描述'
        ];

        const csvContent = [
            headers.join(','),
            ...pods.map(pod => [
                pod.cluster_name || '',
                pod.pod_name,
                pod.namespace,
                pod.node_name || '',
                pod.memory_usage || 0,
                pod.memory_request || 0,
                (pod.memory_req_pct || 0).toFixed(3),
                pod.memory_limit || 0,
                (pod.memory_limit_pct || 0).toFixed(3),
                pod.cpu_usage || 0,
                pod.cpu_request || 0,
                (pod.cpu_req_pct || 0).toFixed(3),
                pod.cpu_limit || 0,
                (pod.cpu_limit_pct || 0).toFixed(3),
                pod.status,
                (pod.issues || []).join('; ')
            ].map(field => `"${field}"`).join(','))
        ].join('\n');

        return csvContent;
    }

    downloadCSV(csvContent, filename) {
        const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
        const link = document.createElement('a');
        
        if (link.download !== undefined) {
            const url = URL.createObjectURL(blob);
            link.setAttribute('href', url);
            link.setAttribute('download', filename);
            link.style.visibility = 'hidden';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        }
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    new ResourceMonitor();
});