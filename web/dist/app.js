"use strict";(()=>{var $=localStorage.getItem("token");$||(window.location.href="/login.html");var L=document.getElementById("themeFloatingBtn"),S=document.getElementById("logoutBtn"),D=document.getElementById("addInvestmentForm"),d=document.getElementById("performanceChart"),T=document.getElementById("metricsGrid"),w=document.getElementById("assetAllocation"),I=document.getElementById("topPerformers"),k=document.getElementById("holdingsTable");L?.addEventListener("click",A);S?.addEventListener("click",()=>{localStorage.removeItem("token"),window.location.href="/login.html"});var M=[{id:"1",name:"Apple Inc.",ticker:"AAPL",quantity:50,purchasePrice:150.25,currentPrice:178.5,purchaseDate:"2024-01-15",type:"stock"},{id:"2",name:"Vanguard S&P 500 ETF",ticker:"VOO",quantity:25,purchasePrice:380,currentPrice:425.75,purchaseDate:"2024-02-10",type:"etf"},{id:"3",name:"Bitcoin",ticker:"BTC",quantity:.5,purchasePrice:45e3,currentPrice:52e3,purchaseDate:"2024-03-05",type:"crypto"},{id:"4",name:"Microsoft Corporation",ticker:"MSFT",quantity:30,purchasePrice:320.5,currentPrice:385.2,purchaseDate:"2024-01-20",type:"stock"},{id:"5",name:"Tesla Inc.",ticker:"TSLA",quantity:15,purchasePrice:245,currentPrice:218.5,purchaseDate:"2024-04-01",type:"stock"},{id:"6",name:"NVIDIA Corporation",ticker:"NVDA",quantity:20,purchasePrice:480,currentPrice:725.3,purchaseDate:"2024-01-25",type:"stock"},{id:"7",name:"Ethereum",ticker:"ETH",quantity:5,purchasePrice:2200,currentPrice:2850,purchaseDate:"2024-03-15",type:"crypto"},{id:"8",name:"iShares Core MSCI Emerging Markets ETF",ticker:"IEMG",quantity:100,purchasePrice:48.5,currentPrice:52.2,purchaseDate:"2024-02-20",type:"etf"},{id:"9",name:"Amazon.com Inc.",ticker:"AMZN",quantity:40,purchasePrice:142,currentPrice:165.8,purchaseDate:"2024-02-05",type:"stock"},{id:"10",name:"US Treasury Bond",ticker:"TLT",quantity:50,purchasePrice:95,currentPrice:97.5,purchaseDate:"2024-01-10",type:"bond"}],l=[];function q(){let t=localStorage.getItem("theme"),n=window.matchMedia("(prefers-color-scheme: dark)").matches;t==="dark"||!t&&n?document.documentElement.classList.add("dark"):document.documentElement.classList.remove("dark"),P()}function P(){let t=document.documentElement.classList.contains("dark"),n=document.getElementById("floatingIconDark"),r=document.getElementById("floatingIconLight");n&&r&&(n.classList.toggle("hidden",!t),r.classList.toggle("hidden",t))}function A(){document.documentElement.classList.contains("dark")?(document.documentElement.classList.remove("dark"),localStorage.setItem("theme","light")):(document.documentElement.classList.add("dark"),localStorage.setItem("theme","dark")),P()}q();window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change",t=>{localStorage.getItem("theme")||(t.matches?document.documentElement.classList.add("dark"):document.documentElement.classList.remove("dark"),P())});function b(t){if(!t)return 0;let n=parseFloat(t.units||t.Units||"0")||0,r=(t.nano||t.Nano||0)/1e9;return n+r}function p(t){if(!t)return 0;let n=parseFloat(t.units||t.Units||"0")||0,r=(t.nano||t.Nano||0)/1e9;return n+r}async function C(){let t=localStorage.getItem("token");if(!t){window.location.href="/login.html";return}try{console.log("\u0417\u0430\u043F\u0440\u0430\u0448\u0438\u0432\u0430\u0435\u043C \u043F\u043E\u0440\u0442\u0444\u0435\u043B\u044C \u0441 \u0442\u043E\u043A\u0435\u043D\u043E\u043C:",t);let n=await fetch("/auth",{method:"POST",headers:{"T-Token":t,"Content-Type":"application/json"}});if(console.log("\u0421\u0442\u0430\u0442\u0443\u0441 \u043E\u0442\u0432\u0435\u0442\u0430 \u043F\u043E\u0440\u0442\u0444\u0435\u043B\u044F:",n.status),n.status===401){console.log("Token invalid or expired"),localStorage.removeItem("token"),window.location.href="/login.html";return}if(!n.ok)throw new Error(`HTTP error! status: ${n.status}`);let r=await n.json();console.log("\u2705 \u0421\u0435\u0440\u0432\u0435\u0440 \u043E\u0442\u0432\u0435\u0442\u0438\u043B:",r);let o;"user_portfolio"in r&&r.user_portfolio?o=r.user_portfolio:o=r,console.log("\u041F\u043E\u0440\u0442\u0444\u0435\u043B\u044C:",o);let i=(o.positions||[]).filter(e=>e.instrumentType?.toLowerCase()!=="currency");console.log(`\u0412\u0441\u0435\u0433\u043E \u043F\u043E\u0437\u0438\u0446\u0438\u0439: ${o.positions?.length||0}, \u043F\u043E\u0441\u043B\u0435 \u0444\u0438\u043B\u044C\u0442\u0440\u0430\u0446\u0438\u0438: ${i.length}`),l=i.map(e=>(console.log("\u041E\u0431\u0440\u0430\u0431\u0430\u0442\u044B\u0432\u0430\u0435\u043C \u043F\u043E\u0437\u0438\u0446\u0438\u044E:",e),{id:e.positionUid||e.PositionUID||Math.random().toString(),name:e.name||e.Name||"Unknown",ticker:e.ticker||e.Ticker||"N/A",type:(e.instrumentType||e.InstrumentType||"stock").toLowerCase(),quantity:b(e.quantity),purchasePrice:p(e.averagePositionPrice),currentPrice:p(e.currentPrice),purchaseDate:new Date().toISOString().split("T")[0]})),window.fullPortfolio=o,y(),x(),console.log("\u2705 \u0418\u043D\u0432\u0435\u0441\u0442\u0438\u0446\u0438\u0438 \u0437\u0430\u0433\u0440\u0443\u0436\u0435\u043D\u044B:",l.length,"\u043F\u043E\u0437\u0438\u0446\u0438\u0439")}catch(n){console.error("\u274C \u041E\u0448\u0438\u0431\u043A\u0430 \u0437\u0430\u0433\u0440\u0443\u0437\u043A\u0438 \u043F\u043E\u0440\u0442\u0444\u0435\u043B\u044F:",n),alert("Failed to load portfolio. Check console for details."),console.log("\u0418\u0441\u043F\u043E\u043B\u044C\u0437\u0443\u0435\u043C \u043C\u043E\u043A\u043E\u0432\u044B\u0435 \u0434\u0430\u043D\u043D\u044B\u0435"),l=[...M],y(),x()}}function y(){localStorage.setItem("investments",JSON.stringify(l))}function g(t){return new Intl.NumberFormat("ru-RU",{style:"currency",currency:"RUB",minimumFractionDigits:2,maximumFractionDigits:2}).format(t)}function B(){let t=window.fullPortfolio,n=0;t?.totalAmountPortfolio?n=p(t.totalAmountPortfolio):n=l.reduce((c,s)=>c+s.quantity*s.currentPrice,0);let r=0;if(t?.expectedYield)r=p(t.expectedYield);else{let c=l.reduce((m,a)=>m+a.quantity*a.purchasePrice,0);r=l.reduce((m,a)=>m+a.quantity*a.currentPrice,0)-c}let o=0,i=t?.positions||[];i.length>0?o=i.reduce((c,s)=>{let m=b(s.quantity),a=p(s.averagePositionPrice);return c+m*a},0):o=l.reduce((c,s)=>c+s.quantity*s.purchasePrice,0);let e=l.filter(c=>c.type!=="currency").length,u=o>0?r/o*100:0;return{totalInvested:o,currentValue:n,totalGain:r,totalGainPercent:u,portfolioSize:e}}function F(){if(!T)return;let t=l.length,{totalInvested:n,currentValue:r,totalGain:o,totalGainPercent:i}=B();T.innerHTML=`
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
            <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                <h3 class="tracking-tight text-sm font-medium">Total Invested</h3>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground"><path d="M19 7V4a1 1 0 0 0-1-1H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v4h-3a2 2 0 0 0 0 4h3a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1"/><path d="M3 5v14a2 2 0 0 0 2 2h15a1 1 0 0 0 1-1v-4"/></svg>
            </div>
            <div class="p-6 pt-0">
                <div class="text-2xl font-bold">${g(n)}</div>
                <p class="text-xs text-muted-foreground">Across ${l.length} ${l.length===1?"investment":"investments"}</p>
            </div>
        </div>
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
            <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                <h3 class="tracking-tight text-sm font-medium">Current Value</h3>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground"><line x1="12" x2="12" y1="2" y2="22"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
            </div>
            <div class="p-6 pt-0">
                <div class="text-2xl font-bold">${g(r)}</div>
                <p class="text-xs text-muted-foreground">Market value of holdings</p>
            </div>
        </div>
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
            <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                <h3 class="tracking-tight text-sm font-medium">Total Gain/Loss</h3>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground"><polyline points="22 7 13.5 15.5 8.5 10.5 2 17"/><polyline points="16 7 22 7 22 13"/></svg>
            </div>
            <div class="p-6 pt-0">
                <div class="text-2xl font-bold ${o>=0?"text-green-600":"text-red-600"}">
                    ${o>=0?"+":""}${g(o)}
                </div>
                <p class="text-xs ${i>=0?"text-green-600":"text-red-600"}">
                    ${i>=0?"+":""}${i.toFixed(2)}% return
                </p>
            </div>
        </div>
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
            <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                <h3 class="tracking-tight text-sm font-medium">Portfolio Size</h3>
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-muted-foreground"><path d="M21.21 15.89A10 10 0 1 1 8 2.83"/><path d="M22 12A10 10 0 0 0 12 2v10z"/></svg>
            </div>
            <div class="p-6 pt-0">
                <div class="text-2xl font-bold">${t}</div>
                <p class="text-xs text-muted-foreground">Active positions</p>
            </div>
        </div>
    `}function E(){if(!d)return;let t=d.getContext("2d");if(!t)return;d.width=d.offsetWidth,d.height=300;let n=[...l].sort((a,h)=>new Date(a.purchaseDate).getTime()-new Date(h.purchaseDate).getTime()),r=0,o=0,i=n.map(a=>(r+=a.quantity*a.purchasePrice,o+=a.quantity*a.currentPrice,{date:a.purchaseDate,invested:r,value:o}));if(i.length===0){t.fillStyle="#888",t.font="16px sans-serif",t.textAlign="center",t.fillText("No data to display",d.width/2,d.height/2);return}let e=40,u=d.width-e*2,c=d.height-e*2,s=Math.max(...i.map(a=>Math.max(a.invested,a.value))),m=0;t.clearRect(0,0,d.width,d.height),t.strokeStyle="#e5e7eb",t.lineWidth=1,t.beginPath(),t.moveTo(e,e),t.lineTo(e,d.height-e),t.lineTo(d.width-e,d.height-e),t.stroke(),t.strokeStyle="#3b82f6",t.lineWidth=2,t.beginPath(),i.forEach((a,h)=>{let f=e+h/(i.length-1)*u,v=d.height-e-(a.invested-m)/(s-m)*c;h===0?t.moveTo(f,v):t.lineTo(f,v)}),t.stroke(),t.strokeStyle="#10b981",t.lineWidth=2,t.beginPath(),i.forEach((a,h)=>{let f=e+h/(i.length-1)*u,v=d.height-e-(a.value-m)/(s-m)*c;h===0?t.moveTo(f,v):t.lineTo(f,v)}),t.stroke(),t.font="12px sans-serif",t.fillStyle="#3b82f6",t.fillRect(e,10,15,15),t.fillStyle="#000",t.fillText("Invested",e+20,22),t.fillStyle="#10b981",t.fillRect(e+120,10,15,15),t.fillStyle="#000",t.fillText("Current Value",e+140,22)}function j(){if(!w)return;let t={},n=0,r=window.fullPortfolio?.positions||[];if(console.log("\u0412\u0441\u0435 \u043F\u043E\u0437\u0438\u0446\u0438\u0438 \u0434\u043B\u044F Asset Allocation:",r.length),r.forEach(e=>{let u=b(e.quantity),c=p(e.currentPrice),s=u*c,m=(e.instrumentType||e.InstrumentType||"other").toLowerCase();t[m]=(t[m]||0)+s,n+=s,console.log(`\u041F\u043E\u0437\u0438\u0446\u0438\u044F: ${e.name||e.Name}, \u0442\u0438\u043F: ${m}, \u0441\u0442\u043E\u0438\u043C\u043E\u0441\u0442\u044C: ${s}`)}),console.log("Allocation \u043F\u043E \u0442\u0438\u043F\u0430\u043C:",t),console.log("\u041E\u0431\u0449\u0430\u044F \u0441\u0442\u043E\u0438\u043C\u043E\u0441\u0442\u044C:",n),n===0){w.innerHTML='<p class="text-sm text-muted-foreground">No data</p>';return}let i=Object.entries(t).sort((e,u)=>u[1]-e[1]).map(([e,u])=>{let c=(u/n*100).toFixed(1),s=e;switch(e){case"stock":s="Stocks";break;case"etf":s="ETFs";break;case"bond":s="Bonds";break;case"crypto":s="Crypto";break;case"currency":s="Currency";break;default:s=e.charAt(0).toUpperCase()+e.slice(1)}return`
            <div class="space-y-1">
                <div class="flex items-center justify-between">
                    <span class="text-sm font-medium">${s}</span>
                    <span class="text-sm text-muted-foreground">${c}%</span>
                </div>
                <div class="w-full bg-muted rounded-full h-2">
                    <div class="bg-primary rounded-full h-2" style="width: ${c}%"></div>
                </div>
            </div>
        `}).join("");w.innerHTML=i}function H(){if(!I)return;let r=l.map(o=>{let i=o.quantity*o.purchasePrice,u=o.quantity*o.currentPrice-i,c=i>0?u/i*100:0;return{...o,gain:u,gainPercent:c}}).sort((o,i)=>i.gainPercent-o.gainPercent).slice(0,5).map(o=>`
        <div class="flex items-center justify-between">
            <div>
                <div class="text-sm font-medium">${o.name}</div>
                <div class="text-xs text-muted-foreground">${o.ticker}</div>
            </div>
            <div class="text-right">
                <div class="text-sm font-bold ${o.gainPercent>=0?"text-green-600":"text-red-600"}">
                    ${o.gainPercent>=0?"+":""}${o.gainPercent.toFixed(2)}%
                </div>
                <div class="text-xs ${o.gainPercent>=0?"text-green-600":"text-red-600"}">
                    ${o.gain>=0?"+":""}${g(o.gain)}
                </div>
            </div>
        </div>
    `).join("");I.innerHTML=r}function N(){if(!k)return;if(l.length===0){k.innerHTML='<tr><td colspan="9" class="p-4 text-center text-muted-foreground">No investments yet</td></tr>';return}let t=l.map(n=>{let r=n.quantity*n.purchasePrice,o=n.quantity*n.currentPrice,i=o-r,e=r>0?i/r*100:0,u=i>=0?'<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline mr-1"><polyline points="22 7 13.5 15.5 8.5 10.5 2 17"/><polyline points="16 7 22 7 22 13"/></svg>':'<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline mr-1"><polyline points="22 17 13.5 8.5 8.5 13.5 2 7"/><polyline points="16 17 22 17 22 11"/></svg>';return`
            <tr class="border-b transition-colors hover:bg-muted/50">
                <td class="p-4 align-middle">${n.name}</td>
                <td class="p-4 align-middle">${n.ticker}</td>
                <td class="p-4 align-middle capitalize">${n.type}</td>
                <td class="p-4 align-middle text-right">${n.quantity}</td>
                <td class="p-4 align-middle text-right">${g(n.purchasePrice)}</td>
                <td class="p-4 align-middle text-right">${g(n.currentPrice)}</td>
                <td class="p-4 align-middle text-right">${g(o)}</td>
                <td class="p-4 align-middle text-right ${i>=0?"text-green-600":"text-red-600"}">
                    ${i>=0?"+":""}${g(i)} (${e>=0?"+":""}${e.toFixed(2)}%)
                    ${u}
                </td>
                <td class="p-4 align-middle text-center">
                    <button onclick="deleteInvestment('${n.id}')" class="text-red-600 hover:text-red-800">Delete</button>
                </td>
            </tr>
        `}).join("");k.innerHTML=t}function x(){F(),E(),j(),H(),N()}document.querySelectorAll(".tab-btn").forEach(t=>{t.addEventListener("click",n=>{let r=n.currentTarget,o=r.dataset.tab;if(!o)return;document.querySelectorAll(".tab-btn").forEach(e=>e.classList.remove("active")),r.classList.add("active"),document.querySelectorAll(".tab-content").forEach(e=>{e.classList.add("hidden")});let i=document.getElementById(`tab-${o}`);i&&i.classList.remove("hidden"),o==="overview"&&E()})});D?.addEventListener("submit",t=>{t.preventDefault();let n=t.target,r=new FormData(n),o={id:Date.now().toString(),name:r.get("name")||"",ticker:r.get("ticker")||"",type:r.get("type")||"other",quantity:parseFloat(r.get("quantity"))||0,purchasePrice:parseFloat(r.get("purchasePrice"))||0,currentPrice:parseFloat(r.get("currentPrice"))||0,purchaseDate:r.get("purchaseDate")||new Date().toISOString().split("T")[0]};l.push(o),y(),x(),n.reset();let i=document.querySelector('[data-tab="holdings"]');i&&i.click()});window.deleteInvestment=t=>{confirm("Are you sure you want to delete this investment?")&&(l=l.filter(n=>n.id!==t),y(),x())};C();})();
