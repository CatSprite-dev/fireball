import type { 
    Investment, 
    Portfolio, 
    Position,
    Quantity,
    Price,
    InvestmentWithGain,
    Metrics,
    DataPoint 
} from './types.ts';

const token = localStorage.getItem('token');
if (!token) {
    window.location.href = '/login.html';
}

// DOM
const themeFloatingBtn = document.getElementById('themeFloatingBtn');
const logoutBtn = document.getElementById('logoutBtn');
const addInvestmentForm = document.getElementById('addInvestmentForm') as HTMLFormElement | null;
const performanceChart = document.getElementById('performanceChart') as HTMLCanvasElement | null;
const metricsGrid = document.getElementById('metricsGrid');
const assetAllocation = document.getElementById('assetAllocation');
const topPerformers = document.getElementById('topPerformers');
const holdingsTable = document.getElementById('holdingsTable');

// Event listeners
themeFloatingBtn?.addEventListener('click', toggleTheme);
logoutBtn?.addEventListener('click', () => {
    localStorage.removeItem('token');
    window.location.href = '/login.html';
});

// Mock data (MAYBE DELETE LATER)
const mockInvestments: Investment[] = [
    {
        id: '1',
        name: 'Apple Inc.',
        ticker: 'AAPL',
        quantity: 50,
        purchasePrice: 150.25,
        currentPrice: 178.50,
        purchaseDate: '2024-01-15',
        type: 'stock',
    },
    {
        id: '2',
        name: 'Vanguard S&P 500 ETF',
        ticker: 'VOO',
        quantity: 25,
        purchasePrice: 380.00,
        currentPrice: 425.75,
        purchaseDate: '2024-02-10',
        type: 'etf',
    },
    {
        id: '3',
        name: 'Bitcoin',
        ticker: 'BTC',
        quantity: 0.5,
        purchasePrice: 45000.00,
        currentPrice: 52000.00,
        purchaseDate: '2024-03-05',
        type: 'crypto',
    },
    {
        id: '4',
        name: 'Microsoft Corporation',
        ticker: 'MSFT',
        quantity: 30,
        purchasePrice: 320.50,
        currentPrice: 385.20,
        purchaseDate: '2024-01-20',
        type: 'stock',
    },
    {
        id: '5',
        name: 'Tesla Inc.',
        ticker: 'TSLA',
        quantity: 15,
        purchasePrice: 245.00,
        currentPrice: 218.50,
        purchaseDate: '2024-04-01',
        type: 'stock',
    },
    {
        id: '6',
        name: 'NVIDIA Corporation',
        ticker: 'NVDA',
        quantity: 20,
        purchasePrice: 480.00,
        currentPrice: 725.30,
        purchaseDate: '2024-01-25',
        type: 'stock',
    },
    {
        id: '7',
        name: 'Ethereum',
        ticker: 'ETH',
        quantity: 5,
        purchasePrice: 2200.00,
        currentPrice: 2850.00,
        purchaseDate: '2024-03-15',
        type: 'crypto',
    },
    {
        id: '8',
        name: 'iShares Core MSCI Emerging Markets ETF',
        ticker: 'IEMG',
        quantity: 100,
        purchasePrice: 48.50,
        currentPrice: 52.20,
        purchaseDate: '2024-02-20',
        type: 'etf',
    },
    {
        id: '9',
        name: 'Amazon.com Inc.',
        ticker: 'AMZN',
        quantity: 40,
        purchasePrice: 142.00,
        currentPrice: 165.80,
        purchaseDate: '2024-02-05',
        type: 'stock',
    },
    {
        id: '10',
        name: 'US Treasury Bond',
        ticker: 'TLT',
        quantity: 50,
        purchasePrice: 95.00,
        currentPrice: 97.50,
        purchaseDate: '2024-01-10',
        type: 'bond',
    },
];

// Our investments
let investments: Investment[] = [];

// Theme switching
function initTheme(): void {
    const savedTheme = localStorage.getItem('theme') as 'dark' | 'light' | null;
    const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    
    if (savedTheme === 'dark' || (!savedTheme && systemDark)) {
        document.documentElement.classList.add('dark');
    } else {
        document.documentElement.classList.remove('dark');
    }
    updateThemeIcon();
}

function updateThemeIcon(): void {
    const isDark = document.documentElement.classList.contains('dark');
    
    const floatingIconDark = document.getElementById('floatingIconDark');
    const floatingIconLight = document.getElementById('floatingIconLight');
    
    if (floatingIconDark && floatingIconLight) {
        floatingIconDark.classList.toggle('hidden', !isDark);
        floatingIconLight.classList.toggle('hidden', isDark);
    }
}

function toggleTheme(): void {
    if (document.documentElement.classList.contains('dark')) {
        document.documentElement.classList.remove('dark');
        localStorage.setItem('theme', 'light');
    } else {
        document.documentElement.classList.add('dark');
        localStorage.setItem('theme', 'dark');
    }
    updateThemeIcon();
}

initTheme();

// System theme listener
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e: MediaQueryListEvent) => {
    if (!localStorage.getItem('theme')) {
        if (e.matches) {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
        updateThemeIcon();
    }
});

// Additional parsing functions Tinkoff API
function parseQuantity(quantity?: Quantity): number {
    if (!quantity) return 0;
    const units = parseFloat(quantity.units || quantity.Units || '0') || 0;
    const nano = (quantity.nano || quantity.Nano || 0) / 1e9;
    return units + nano;
}

function parsePrice(price?: Price): number {
    if (!price) return 0;
    const units = parseFloat(price.units || price.Units || '0') || 0;
    const nano = (price.nano || price.Nano || 0) / 1e9;
    return units + nano;
}

// Загружаем данные с бекенда
async function loadInvestments(): Promise<void> {
    const token = localStorage.getItem('token');
    
    if (!token) {
        window.location.href = '/login.html';
        return;
    }
    
    try {
        console.log("Запрашиваем портфель с токеном:", token);
        
        const response = await fetch("/auth", {
            method: 'POST',
            headers: {
                "T-Token": token,
                "Content-Type": "application/json"
            }
        });
        
        console.log("Статус ответа портфеля:", response.status);
        
        if (response.status === 401) {
            console.log('Token invalid or expired');
            localStorage.removeItem('token');
            window.location.href = '/login.html';
            return;
        }
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json() as { user_portfolio?: Portfolio } | Portfolio;
        console.log("✅ Сервер ответил:", data);
        
        // Извлекаем портфель
        let portfolio: Portfolio;
        if ('user_portfolio' in data && data.user_portfolio) {
            portfolio = data.user_portfolio;
        } else {
            portfolio = data as Portfolio;
        }
        
        console.log("Портфель:", portfolio);
        
        // Фильтруем позиции, исключая валюты
        const filteredPositions = (portfolio.positions || []).filter(pos => 
            pos.instrumentType?.toLowerCase() !== 'currency'
        );

        console.log(`Всего позиций: ${portfolio.positions?.length || 0}, после фильтрации: ${filteredPositions.length}`);

        investments = filteredPositions.map(pos => {
            console.log("Обрабатываем позицию:", pos);
            return {
                id: pos.positionUid || pos.PositionUID || Math.random().toString(),
                name: pos.name || pos.Name || 'Unknown',
                ticker: pos.ticker || pos.Ticker || 'N/A',
                type: (pos.instrumentType || pos.InstrumentType || 'stock').toLowerCase(),
                quantity: parseQuantity(pos.quantity),
                purchasePrice: parsePrice(pos.averagePositionPrice),
                currentPrice: parsePrice(pos.currentPrice),
                purchaseDate: new Date().toISOString().split('T')[0]
            };
        });
        
        window.fullPortfolio = portfolio;
        saveInvestments();
        renderAll();
        
        console.log("✅ Инвестиции загружены:", investments.length, "позиций");
        
    } catch (error) {
        console.error('❌ Ошибка загрузки портфеля:', error);
        alert('Failed to load portfolio. Check console for details.');
        
        console.log("Используем моковые данные");
        investments = [...mockInvestments];
        saveInvestments();
        renderAll();
    }
}

function saveInvestments(): void {
    localStorage.setItem('investments', JSON.stringify(investments));
}

function formatCurrency(num: number): string {
    return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: 'RUB',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
    }).format(num);
}

function calculateMetrics(): Metrics {
    const portfolio = window.fullPortfolio;
    
    // Current Value
    let currentValue = 0;
    if (portfolio?.totalAmountPortfolio) {
        currentValue = parsePrice(portfolio.totalAmountPortfolio);
    } else {
        currentValue = investments.reduce((sum, inv) => sum + inv.quantity * inv.currentPrice, 0);
    }
    
    // Total Gain/Loss
    let totalGain = 0;
    if (portfolio?.expectedYield) {
        totalGain = parsePrice(portfolio.expectedYield);
    } else {
        const totalInvested = investments.reduce((sum, inv) => sum + inv.quantity * inv.purchasePrice, 0);
        const currentVal = investments.reduce((sum, inv) => sum + inv.quantity * inv.currentPrice, 0);
        totalGain = currentVal - totalInvested;
    }
    
    // Total Invested
    let totalInvested = 0;
    const allPositions = portfolio?.positions || [];
    if (allPositions.length > 0) {
        totalInvested = allPositions.reduce((sum, pos) => {
            const quantity = parseQuantity(pos.quantity);
            const avgPrice = parsePrice(pos.averagePositionPrice);
            return sum + quantity * avgPrice;
        }, 0);
    } else {
        totalInvested = investments.reduce((sum, inv) => sum + inv.quantity * inv.purchasePrice, 0);
    }
    
    const portfolioSize = investments.filter(inv => inv.type !== 'currency').length;
    const totalGainPercent = totalInvested > 0 ? (totalGain / totalInvested) * 100 : 0;
    
    return { 
        totalInvested, 
        currentValue, 
        totalGain, 
        totalGainPercent,
        portfolioSize 
    };
}

function renderMetrics(): void {
    if (!metricsGrid) return;
    
    const portfolioSize = investments.length;
    const { totalInvested, currentValue, totalGain, totalGainPercent } = calculateMetrics();
    
    metricsGrid.innerHTML = `
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
            <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                <h3 class="tracking-tight text-sm font-medium">Total Invested</h3>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground"><path d="M19 7V4a1 1 0 0 0-1-1H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v4h-3a2 2 0 0 0 0 4h3a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1"/><path d="M3 5v14a2 2 0 0 0 2 2h15a1 1 0 0 0 1-1v-4"/></svg>
            </div>
            <div class="p-6 pt-0">
                <div class="text-2xl font-bold">${formatCurrency(totalInvested)}</div>
                <p class="text-xs text-muted-foreground">Across ${investments.length} ${investments.length === 1 ? 'investment' : 'investments'}</p>
            </div>
        </div>
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
            <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                <h3 class="tracking-tight text-sm font-medium">Current Value</h3>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground"><line x1="12" x2="12" y1="2" y2="22"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
            </div>
            <div class="p-6 pt-0">
                <div class="text-2xl font-bold">${formatCurrency(currentValue)}</div>
                <p class="text-xs text-muted-foreground">Market value of holdings</p>
            </div>
        </div>
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
            <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                <h3 class="tracking-tight text-sm font-medium">Total Gain/Loss</h3>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground"><polyline points="22 7 13.5 15.5 8.5 10.5 2 17"/><polyline points="16 7 22 7 22 13"/></svg>
            </div>
            <div class="p-6 pt-0">
                <div class="text-2xl font-bold ${totalGain >= 0 ? 'text-green-600' : 'text-red-600'}">
                    ${totalGain >= 0 ? '+' : ''}${formatCurrency(totalGain)}
                </div>
                <p class="text-xs ${totalGainPercent >= 0 ? 'text-green-600' : 'text-red-600'}">
                    ${totalGainPercent >= 0 ? '+' : ''}${totalGainPercent.toFixed(2)}% return
                </p>
            </div>
        </div>
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
            <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                <h3 class="tracking-tight text-sm font-medium">Portfolio Size</h3>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground"><path d="M21.21 15.89A10 10 0 1 1 8 2.83"/><path d="M22 12A10 10 0 0 0 12 2v10z"/></svg>
            </div>
            <div class="p-6 pt-0">
                <div class="text-2xl font-bold">${portfolioSize}</div>
                <p class="text-xs text-muted-foreground">Active positions</p>
            </div>
        </div>
    `;
}

function renderChart(): void {
    if (!performanceChart) return;
    
    const ctx = performanceChart.getContext('2d');
    if (!ctx) return;
    
    // Set canvas dimensions
    performanceChart.width = performanceChart.offsetWidth;
    performanceChart.height = 300;
    
    // Sort by date
    const sorted = [...investments].sort((a, b) => 
        new Date(a.purchaseDate).getTime() - new Date(b.purchaseDate).getTime()
    );
    
    // Calculate cumulative values
    let cumulativeInvested = 0;
    let cumulativeValue = 0;
    const dataPoints: DataPoint[] = sorted.map(inv => {
        cumulativeInvested += inv.quantity * inv.purchasePrice;
        cumulativeValue += inv.quantity * inv.currentPrice;
        return { date: inv.purchaseDate, invested: cumulativeInvested, value: cumulativeValue };
    });

    if (dataPoints.length === 0) {
        ctx.fillStyle = '#888';
        ctx.font = '16px sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText('No data to display', performanceChart.width / 2, performanceChart.height / 2);
        return;
    }

    const padding = 40;
    const chartWidth = performanceChart.width - padding * 2;
    const chartHeight = performanceChart.height - padding * 2;
    
    const maxValue = Math.max(...dataPoints.map(d => Math.max(d.invested, d.value)));
    const minValue = 0;
    
    // Clear canvas
    ctx.clearRect(0, 0, performanceChart.width, performanceChart.height);
    
    // Draw axes
    ctx.strokeStyle = '#e5e7eb';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(padding, padding);
    ctx.lineTo(padding, performanceChart.height - padding);
    ctx.lineTo(performanceChart.width - padding, performanceChart.height - padding);
    ctx.stroke();
    
    // Draw invested line
    ctx.strokeStyle = '#3b82f6';
    ctx.lineWidth = 2;
    ctx.beginPath();
    dataPoints.forEach((point, i) => {
        const x = padding + (i / (dataPoints.length - 1)) * chartWidth;
        const y = performanceChart.height - padding - ((point.invested - minValue) / (maxValue - minValue)) * chartHeight;
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
    });
    ctx.stroke();
    
    // Draw current value line
    ctx.strokeStyle = '#10b981';
    ctx.lineWidth = 2;
    ctx.beginPath();
    dataPoints.forEach((point, i) => {
        const x = padding + (i / (dataPoints.length - 1)) * chartWidth;
        const y = performanceChart.height - padding - ((point.value - minValue) / (maxValue - minValue)) * chartHeight;
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
    });
    ctx.stroke();
    
    // Legend
    ctx.font = '12px sans-serif';
    ctx.fillStyle = '#3b82f6';
    ctx.fillRect(padding, 10, 15, 15);
    ctx.fillStyle = '#000';
    ctx.fillText('Invested', padding + 20, 22);
    
    ctx.fillStyle = '#10b981';
    ctx.fillRect(padding + 120, 10, 15, 15);
    ctx.fillStyle = '#000';
    ctx.fillText('Current Value', padding + 140, 22);
}

function renderAssetAllocation(): void {
    if (!assetAllocation) return;
    
    const allocation: Record<string, number> = {};
    let total = 0;
    
    const allPositions = window.fullPortfolio?.positions || [];
    
    console.log("Все позиции для Asset Allocation:", allPositions.length);
    
    allPositions.forEach(pos => {
        const quantity = parseQuantity(pos.quantity);
        const currentPrice = parsePrice(pos.currentPrice);
        const value = quantity * currentPrice;
        
        const type = (pos.instrumentType || pos.InstrumentType || 'other').toLowerCase();
        
        allocation[type] = (allocation[type] || 0) + value;
        total += value;
        
        console.log(`Позиция: ${pos.name || pos.Name}, тип: ${type}, стоимость: ${value}`);
    });
    
    console.log("Allocation по типам:", allocation);
    console.log("Общая стоимость:", total);
    
    if (total === 0) {
        assetAllocation.innerHTML = '<p class="text-sm text-muted-foreground">No data</p>';
        return;
    }
    
    const sortedTypes = Object.entries(allocation).sort((a, b) => b[1] - a[1]);
    
    const html = sortedTypes.map(([type, value]) => {
        const percent = (value / total * 100).toFixed(1);
        
        let typeName = type;
        switch(type) {
            case 'stock':
                typeName = 'Stocks';
                break;
            case 'etf':
                typeName = 'ETFs';
                break;
            case 'bond':
                typeName = 'Bonds';
                break;
            case 'crypto':
                typeName = 'Crypto';
                break;
            case 'currency':
                typeName = 'Currency';
                break;
            default:
                typeName = type.charAt(0).toUpperCase() + type.slice(1);
        }
        
        return `
            <div class="space-y-1">
                <div class="flex items-center justify-between">
                    <span class="text-sm font-medium">${typeName}</span>
                    <span class="text-sm text-muted-foreground">${percent}%</span>
                </div>
                <div class="w-full bg-muted rounded-full h-2">
                    <div class="bg-primary rounded-full h-2" style="width: ${percent}%"></div>
                </div>
            </div>
        `;
    }).join('');
    
    assetAllocation.innerHTML = html;
}

function renderTopPerformers(): void {
    if (!topPerformers) return;
    
    const withGains: InvestmentWithGain[] = investments.map(inv => {
        const invested = inv.quantity * inv.purchasePrice;
        const current = inv.quantity * inv.currentPrice;
        const gain = current - invested;
        const gainPercent = invested > 0 ? (gain / invested) * 100 : 0;
        return { ...inv, gain, gainPercent };
    });
    
    const top = withGains.sort((a, b) => b.gainPercent - a.gainPercent).slice(0, 5);
    
    const html = top.map(inv => `
        <div class="flex items-center justify-between">
            <div>
                <div class="text-sm font-medium">${inv.name}</div>
                <div class="text-xs text-muted-foreground">${inv.ticker}</div>
            </div>
            <div class="text-right">
                <div class="text-sm font-bold ${inv.gainPercent >= 0 ? 'text-green-600' : 'text-red-600'}">
                    ${inv.gainPercent >= 0 ? '+' : ''}${inv.gainPercent.toFixed(2)}%
                </div>
                <div class="text-xs ${inv.gainPercent >= 0 ? 'text-green-600' : 'text-red-600'}">
                    ${inv.gain >= 0 ? '+' : ''}${formatCurrency(inv.gain)}
                </div>
            </div>
        </div>
    `).join('');
    
    topPerformers.innerHTML = html;
}

function renderHoldings(): void {
    if (!holdingsTable) return;
    
    if (investments.length === 0) {
        holdingsTable.innerHTML = '<tr><td colspan="9" class="p-4 text-center text-muted-foreground">No investments yet</td></tr>';
        return;
    }
    
    const html = investments.map(inv => {
        const invested = inv.quantity * inv.purchasePrice;
        const current = inv.quantity * inv.currentPrice;
        const gain = current - invested;
        const gainPercent = invested > 0 ? (gain / invested) * 100 : 0;
        
        const gainIcon = gain >= 0 ? 
            '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline mr-1"><polyline points="22 7 13.5 15.5 8.5 10.5 2 17"/><polyline points="16 7 22 7 22 13"/></svg>' :
            '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline mr-1"><polyline points="22 17 13.5 8.5 8.5 13.5 2 7"/><polyline points="16 17 22 17 22 11"/></svg>';
        
        return `
            <tr class="border-b transition-colors hover:bg-muted/50">
                <td class="p-4 align-middle">${inv.name}</td>
                <td class="p-4 align-middle">${inv.ticker}</td>
                <td class="p-4 align-middle capitalize">${inv.type}</td>
                <td class="p-4 align-middle text-right">${inv.quantity}</td>
                <td class="p-4 align-middle text-right">${formatCurrency(inv.purchasePrice)}</td>
                <td class="p-4 align-middle text-right">${formatCurrency(inv.currentPrice)}</td>
                <td class="p-4 align-middle text-right">${formatCurrency(current)}</td>
                <td class="p-4 align-middle text-right ${gain >= 0 ? 'text-green-600' : 'text-red-600'}">
                    ${gain >= 0 ? '+' : ''}${formatCurrency(gain)} (${gainPercent >= 0 ? '+' : ''}${gainPercent.toFixed(2)}%)
                    ${gainIcon}
                </td>
                <td class="p-4 align-middle text-center">
                    <button onclick="deleteInvestment('${inv.id}')" class="text-red-600 hover:text-red-800">Delete</button>
                </td>
            </tr>
        `;
    }).join('');
    
    holdingsTable.innerHTML = html;
}

function renderAll(): void {
    renderMetrics();
    renderChart();
    renderAssetAllocation();
    renderTopPerformers();
    renderHoldings();
}

// Tab switching
document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
        const target = e.currentTarget as HTMLElement;
        const tab = target.dataset.tab;
        if (!tab) return;
        
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        target.classList.add('active');
        
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.add('hidden');
        });
        
        const tabElement = document.getElementById(`tab-${tab}`);
        if (tabElement) {
            tabElement.classList.remove('hidden');
        }
        
        if (tab === 'overview') {
            renderChart();
        }
    });
});

// Add investment form
addInvestmentForm?.addEventListener('submit', (e: Event) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const formData = new FormData(form);
    
    const newInvestment: Investment = {
        id: Date.now().toString(),
        name: formData.get('name') as string || '',
        ticker: formData.get('ticker') as string || '',
        type: formData.get('type') as string || 'other',
        quantity: parseFloat(formData.get('quantity') as string) || 0,
        purchasePrice: parseFloat(formData.get('purchasePrice') as string) || 0,
        currentPrice: parseFloat(formData.get('currentPrice') as string) || 0,
        purchaseDate: formData.get('purchaseDate') as string || new Date().toISOString().split('T')[0],
    };
    
    investments.push(newInvestment);
    saveInvestments();
    renderAll();
    form.reset();
    
    // Switch to holdings tab
    const holdingsTab = document.querySelector('[data-tab="holdings"]') as HTMLElement;
    if (holdingsTab) {
        holdingsTab.click();
    }
});

// Delete investment
window.deleteInvestment = (id: string): void => {
    if (confirm('Are you sure you want to delete this investment?')) {
        investments = investments.filter(inv => inv.id !== id);
        saveInvestments();
        renderAll();
    }
};

// Initialize
loadInvestments();