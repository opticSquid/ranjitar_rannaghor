import { createSignal, onMount, For, Show } from 'solid-js';
import axios from 'axios';
import { User, BillReport } from '../types';
import { Download, Search } from 'lucide-solid';
import { jsPDF } from 'jspdf';
import html2canvas from 'html2canvas';
import { useI18n } from '../i18n';

import { globalUserTrie, loadUsers } from '../store/userStore';

const Billing = () => {
    const { t } = useI18n();
    const [selectedUser, setSelectedUser] = createSignal<string>('');
    const [startDate, setStartDate] = createSignal(new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split('T')[0]);
    const [endDate, setEndDate] = createSignal(new Date().toISOString().split('T')[0]);
    const [report, setReport] = createSignal<BillReport | null>(null);
    const [isLoading, setIsLoading] = createSignal(false);
    
    const [searchQuery, setSearchQuery] = createSignal('');
    const [suggestions, setSuggestions] = createSignal<User[]>([]);
    const [showSuggestions, setShowSuggestions] = createSignal(false);

    onMount(loadUsers);

    const handleSearchInput = (e: any) => {
        const val = e.currentTarget.value;
        setSearchQuery(val);
        
        if (selectedUser()) {
            setSelectedUser(''); 
        }
        
        if (val.trim().length >= 3) {
            const results = globalUserTrie().search(val.trim(), 5);
            setSuggestions(results);
            setShowSuggestions(true);
        } else {
            setSuggestions([]);
            setShowSuggestions(false);
        }
    };

    const handleSelectUser = (u: User) => {
        setSelectedUser(u.user_id.toString());
        setSearchQuery(u.name);
        setShowSuggestions(false);
    };

    const handleSearchBlur = () => {
        setTimeout(() => setShowSuggestions(false), 200);
    };

    const generateReport = async () => {
        if (!selectedUser()) return;
        setIsLoading(true);
        try {
            const res = await axios.get(`/api/reports/bill?user_id=${selectedUser()}&start_date=${startDate()}&end_date=${endDate()}`);
            setReport(res.data);
        } catch (err) {
            alert('Failed to generate report');
        } finally {
            setIsLoading(false);
        }
    };

    const exportPDF = async () => {
        const element = document.getElementById('bill-content');
        if (!element) return;

        const originalScroll = window.scrollY;
        window.scrollTo(0, 0);

        await new Promise(resolve => setTimeout(resolve, 500));

        try {
            const canvas = await html2canvas(element, {
                scale: 2,
                backgroundColor: '#ffffff',
                logging: false,
                useCORS: true,
                allowTaint: false,
                windowWidth: element.scrollWidth,
                windowHeight: element.scrollHeight,
            });

            const imgData = canvas.toDataURL('image/png');
            const pdf = new jsPDF('p', 'mm', 'a4');
            const pdfWidth = pdf.internal.pageSize.getWidth();
            const pdfHeight = pdf.internal.pageSize.getHeight();
            const imgHeight = (canvas.height * pdfWidth) / canvas.width;

            let heightLeft = imgHeight;
            let position = 0;

            pdf.addImage(imgData, 'PNG', 0, position, pdfWidth, imgHeight);
            heightLeft -= pdfHeight;

            while (heightLeft > 0) {
                position = position - pdfHeight;
                pdf.addPage();
                pdf.addImage(imgData, 'PNG', 0, position, pdfWidth, imgHeight);
                heightLeft -= pdfHeight;
            }

            pdf.save(`${report()?.user.name}_${startDate()}_to_${endDate()}.pdf`);
        } catch (err) {
            console.error('PDF Export Error:', err);
            alert('Failed to export PDF.');
        } finally {
            window.scrollTo(0, originalScroll);
        }
    };

    return (
        <div class="space-y-6 pb-24 animate-in">
            <header>
                <h2 class="text-2xl font-bold text-slate-800">{t('billingAndReports')}</h2>
                <p class="text-slate-500 font-medium text-sm mt-1">{t('generateAndExport')}</p>
            </header>

            <div class="card bg-white border border-slate-200 p-5 space-y-4">
                
                {/* Search */}
                <div>
                    <label class="block text-sm font-bold text-slate-700 mb-1">{t('customer')}</label>
                    <div class="relative z-40">
                        <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                            <Search class="text-slate-400" size={24} />
                        </div>
                        <input
                            type="text"
                            class="input-large pl-12"
                            placeholder={t('selectCustomer') || "Search Customer Name..."}
                            value={searchQuery()}
                            onInput={handleSearchInput}
                            onFocus={() => { if (searchQuery().trim().length >= 3) setShowSuggestions(true); }}
                            onBlur={handleSearchBlur}
                        />
                        {showSuggestions() && (
                            <ul class="absolute top-full left-0 right-0 mt-1 bg-white rounded-xl shadow-xl border border-slate-200 overflow-hidden">
                                <For each={suggestions()}>
                                    {(user) => (
                                        <li 
                                            class="p-4 cursor-pointer hover:bg-slate-50 border-b border-slate-100 last:border-0"
                                            onClick={() => handleSelectUser(user)}
                                        >
                                            <span class="font-bold text-lg text-slate-900 block">{user.name}</span>
                                            <span class="text-sm text-slate-500">Room {user.room_no}</span>
                                        </li>
                                    )}
                                </For>
                                {suggestions().length === 0 && (
                                    <li class="p-6 text-center text-slate-500 font-medium">No customers found</li>
                                )}
                            </ul>
                        )}
                    </div>
                </div>

                {/* Dates */}
                <div class="grid grid-cols-2 gap-4 pt-2">
                    <div>
                        <label class="block text-sm font-bold text-slate-700 mb-1">{t('from')}</label>
                        <input
                            type="date"
                            class="input-large !h-14 !text-sm"
                            value={startDate()}
                            onInput={e => setStartDate(e.currentTarget.value)}
                        />
                    </div>
                    <div>
                        <label class="block text-sm font-bold text-slate-700 mb-1">{t('to')}</label>
                        <input
                            type="date"
                            class="input-large !h-14 !text-sm"
                            value={endDate()}
                            onInput={e => setEndDate(e.currentTarget.value)}
                        />
                    </div>
                </div>

                <button
                    onClick={generateReport}
                    disabled={isLoading() || !selectedUser()}
                    class={`btn btn-primary w-full mt-4 text-xl ${(!selectedUser() || isLoading()) ? 'opacity-50' : ''}`}
                >
                    {isLoading() ? t('generating') : t('generateBill')}
                </button>
            </div>

            <Show when={report()}>
                <div class="animate-in slide-in-from-bottom">
                    <button 
                        onClick={exportPDF} 
                        class="btn btn-success w-full mb-6 text-lg shadow-lg shadow-green-500/30"
                    >
                        <Download size={24} /> {t('downloadPdf')}
                    </button>

                    <div id="bill-content" class="bg-white text-slate-800 p-6 sm:p-12 rounded-xl shadow-md max-w-4xl mx-auto border border-slate-200">
                        {/* Bill Header */}
                        <div class="flex justify-between items-start border-b-2 border-slate-200 pb-6 mb-6">
                            <div>
                                <div class="flex items-center gap-2 mb-2">
                                    <div class="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center">
                                        <span class="text-white font-black text-xl">R</span>
                                    </div>
                                    <h1 class="text-xl font-black text-slate-900 tracking-tight">{t('adminSubtitle')}</h1>
                                </div>
                                <p class="text-slate-500 text-xs">{t('homeCookedSubtitle')}</p>
                            </div>
                            <div class="text-right">
                                <h2 class="text-2xl font-black text-slate-800 uppercase tracking-tighter">{t('invoice')}</h2>
                                <p class="text-slate-500 text-xs mt-1 font-bold">#{report()?.user.user_id}-{new Date().getTime().toString().slice(-6)}</p>
                                <div class="mt-2 text-xs font-bold text-blue-700 bg-blue-50 px-3 py-1 rounded-full inline-block">
                                    {new Date(startDate()).toLocaleDateString()} - {new Date(endDate()).toLocaleDateString()}
                                </div>
                            </div>
                        </div>

                        {/* Bill Info */}
                        <div class="grid grid-cols-2 gap-6 mb-8 text-sm">
                            <div>
                                <h4 class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1">{t('billTo')}</h4>
                                <p class="text-lg font-black text-slate-800">{report()?.user.name}</p>
                                <p class="text-slate-600 font-medium">{report()?.user.mobile_no}</p>
                                <p class="text-slate-600 font-medium">Room {report()?.user.room_no}</p>
                            </div>
                            <div class="text-right">
                                <h4 class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1">{t('summary')}</h4>
                                <div class="space-y-1 text-sm font-medium">
                                    <p class="text-slate-500">{t('openingBalance')} <span class="text-slate-800 font-bold ml-2">₹{report()?.opening_balance.toFixed(2)}</span></p>
                                    <p class="text-slate-500">{t('totalBillable')} <span class="text-slate-800 font-bold ml-2">₹{report()?.total_spent.toFixed(2)}</span></p>
                                    <p class="text-slate-500">{t('recharges')} <span class="text-slate-800 font-bold ml-2">₹{report()?.total_recharges.toFixed(2)}</span></p>
                                </div>
                            </div>
                        </div>

                        {/* Bill Table */}
                        <div class="overflow-x-auto">
                            <table class="w-full mb-8 text-sm">
                                <thead>
                                    <tr class="bg-slate-100 text-slate-600 text-left">
                                        <th class="py-2 px-3 font-bold text-[10px] uppercase tracking-wider">{t('date')}</th>
                                        <th class="py-2 px-3 font-bold text-[10px] uppercase tracking-wider">{t('meal')}</th>
                                        <th class="py-2 px-3 font-bold text-[10px] uppercase tracking-wider">{t('details')}</th>
                                        <th class="py-2 px-3 font-bold text-[10px] uppercase tracking-wider text-right">{t('amount')}</th>
                                    </tr>
                                </thead>
                                <tbody class="divide-y divide-slate-100">
                                    <For each={report()?.logs}>
                                        {(log) => (
                                            <tr class="text-slate-800">
                                                <td class="py-3 px-3">{new Date(log.log_date).toLocaleDateString()}</td>
                                                <td class="py-3 px-3 font-bold capitalize text-slate-600">{log.meal_type}</td>
                                                <td class="py-3 px-3">
                                                    {log.has_main_meal ? (
                                                        log.is_special ? (
                                                            <span class="font-bold text-purple-700">{log.special_dish_name} </span>
                                                        ) : (
                                                            t('standard')
                                                        )
                                                    ) : (
                                                        <span class="text-slate-500 italic">{t('aLaCarte')}</span>
                                                    )}
                                                    {log.extra_rice_qty > 0 && <span class="text-slate-500 text-xs ml-1">+Rice({log.extra_rice_qty}) </span>}
                                                    {log.extra_roti_qty > 0 && <span class="text-slate-500 text-xs ml-1">+Roti({log.extra_roti_qty}) </span>}
                                                    {log.extra_chicken_qty > 0 && <span class="text-slate-500 text-xs ml-1">+Chicken({log.extra_chicken_qty}) </span>}
                                                    {log.extra_fish_qty > 0 && <span class="text-slate-500 text-xs ml-1">+Fish({log.extra_fish_qty}) </span>}
                                                </td>
                                                <td class="py-3 px-3 font-black text-right">₹{log.total_cost.toFixed(2)}</td>
                                            </tr>
                                        )}
                                    </For>
                                </tbody>
                            </table>
                        </div>

                        {/* Bill Totals */}
                        <div class="flex justify-end pt-4 border-t-2 border-slate-200">
                            <div class="w-64 space-y-3">
                                <div class="flex justify-between text-slate-600 font-bold">
                                    <span>{t('subtotal')}</span>
                                    <span>₹{report()?.total_spent.toFixed(2)}</span>
                                </div>
                                <div class="flex justify-between text-xl font-black text-slate-900 border-t border-slate-200 pt-3">
                                    <span>{t('total')}</span>
                                    <span>₹{report()?.total_spent.toFixed(2)}</span>
                                </div>
                                <div class={`p-3 rounded-xl mt-3 flex justify-between items-center ${report()!.closing_balance < 0 ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200'}`}>
                                    <span class="text-[10px] font-black uppercase tracking-wider">{t('balanceStatus')}</span>
                                    <span class="font-black text-lg">₹{report()?.closing_balance.toFixed(2)}</span>
                                </div>
                            </div>
                        </div>

                        <div class="mt-12 text-center text-slate-400">
                            <p class="text-[10px] font-bold uppercase tracking-[0.2em] mb-1">{t('generatedVia')}</p>
                            <p class="text-[10px]">{t('thankYou')}</p>
                        </div>
                    </div>
                </div>
            </Show>
        </div>
    );
};

export default Billing;
