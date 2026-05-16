import { createSignal, onMount, For, Show } from 'solid-js';
import axios from 'axios';
import { User, BillReport } from '../types';
import { Download, FileText, Calendar, Search } from 'lucide-solid';
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

        // Ensure we are at the top of the element to capture it fully
        const originalScroll = window.scrollY;
        window.scrollTo(0, 0);

        // Wait a bit to ensure any animations or renders are finished
        await new Promise(resolve => setTimeout(resolve, 500));

        try {
            const canvas = await html2canvas(element, {
                scale: 2,
                backgroundColor: '#ffffff',
                logging: true,
                useCORS: true,
                allowTaint: false,
                windowWidth: element.scrollWidth,
                windowHeight: element.scrollHeight,
                onclone: (clonedDoc) => {
                    const clonedElement = clonedDoc.getElementById('bill-content');
                    if (clonedElement) {
                        clonedElement.style.transform = 'none';
                        clonedElement.style.animation = 'none';
                        clonedElement.style.opacity = '1';
                    }
                }
            });

            const imgData = canvas.toDataURL('image/png');
            const pdf = new jsPDF('p', 'mm', 'a4');
            const pdfWidth = pdf.internal.pageSize.getWidth();
            const pdfHeight = pdf.internal.pageSize.getHeight();
            const imgHeight = (canvas.height * pdfWidth) / canvas.width;

            let heightLeft = imgHeight;
            let position = 0;

            // Add the first page
            pdf.addImage(imgData, 'PNG', 0, position, pdfWidth, imgHeight);
            heightLeft -= pdfHeight;

            // Add subsequent pages if content overflows
            while (heightLeft > 0) {
                position = position - pdfHeight; // Move image up by one page height
                pdf.addPage();
                pdf.addImage(imgData, 'PNG', 0, position, pdfWidth, imgHeight);
                heightLeft -= pdfHeight;
            }

            pdf.save(`${report()?.user.name}_${startDate()}_to_${endDate()}.pdf`);
            console.log('PDF exported successfully');
        } catch (err) {
            console.error('PDF Export Error:', err);
            alert('Failed to export PDF. Check console for details.');
        } finally {
            window.scrollTo(0, originalScroll);
        }
    };

    return (
        <div class="space-y-8 animate-in fade-in duration-700">
            <header>
                <h2 class="text-3xl font-bold text-white">{t('billingAndReports')}</h2>
                <p class="text-text-dim mt-2">{t('generateAndExport')}</p>
            </header>

            <div class="glass p-6 grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
                <div class="md:col-span-1">
                    <label class="block text-xs font-bold uppercase tracking-widest text-text-dim mb-2">{t('customer')}</label>
                    <div class="relative">
                        <input
                            type="text"
                            class="input bg-surface w-full"
                            placeholder={t('selectCustomer') || "Search Customer..."}
                            value={searchQuery()}
                            onInput={handleSearchInput}
                            onFocus={() => { if (searchQuery().trim().length >= 3) setShowSuggestions(true); }}
                            onBlur={handleSearchBlur}
                        />
                        {showSuggestions() && (
                            <ul class="absolute z-50 top-[calc(100%+4px)] left-0 right-0 bg-[var(--md-sys-color-surface-container)] rounded-2xl shadow-xl border border-[var(--md-sys-color-outline-variant)] overflow-hidden">
                                <For each={suggestions()}>
                                    {(user) => (
                                        <li 
                                            class="px-5 py-3 cursor-pointer hover:bg-[var(--md-sys-color-surface-container-high)] border-b border-[var(--md-sys-color-outline-variant)] last:border-0 flex justify-between items-center transition-colors"
                                            onClick={() => handleSelectUser(user)}
                                        >
                                            <span class="font-bold text-[var(--md-sys-color-on-surface)]">{user.name}</span>
                                            <span class="text-xs font-bold text-[var(--md-sys-color-on-surface-variant)] bg-[var(--md-sys-color-surface-container-highest)] px-2.5 py-1 rounded-lg tracking-wide">
                                                {user.role === 'admin' ? 'Admin' : 'User'}
                                            </span>
                                        </li>
                                    )}
                                </For>
                                {suggestions().length === 0 && (
                                    <li class="px-5 py-6 text-sm text-[var(--md-sys-color-on-surface-variant)] text-center font-medium">
                                        No matching customers found
                                    </li>
                                )}
                            </ul>
                        )}
                        <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-text-dim">
                            <Search size={18} />
                        </div>
                    </div>
                </div>
                <div>
                    <label class="block text-xs font-bold uppercase tracking-widest text-text-dim mb-2">{t('from')}</label>
                    <input
                        type="date"
                        class="input"
                        value={startDate()}
                        onInput={e => {
                            const val = e.currentTarget.value;
                            if (val) setStartDate(val);
                        }}
                    />
                </div>
                <div>
                    <label class="block text-xs font-bold uppercase tracking-widest text-text-dim mb-2">{t('to')}</label>
                    <input
                        type="date"
                        class="input"
                        value={endDate()}
                        onInput={e => {
                            const val = e.currentTarget.value;
                            if (val) setEndDate(val);
                        }}
                    />
                </div>
                <button
                    onClick={generateReport}
                    disabled={isLoading() || !selectedUser()}
                    class="btn btn-primary w-full h-[46px]"
                >
                    {isLoading() ? t('generating') : t('generateBill')}
                </button>
            </div>

            <Show when={report()}>
                <div class="animate-in slide-in-from-bottom duration-500">
                    <div class="flex justify-end mb-4">
                        <button onClick={exportPDF} class="bg-[var(--md-sys-color-secondary-container)] text-[var(--md-sys-color-on-secondary-container)] hover:brightness-95 active:scale-95 flex items-center gap-2 rounded-xl px-6 py-3 font-medium transition-all shadow-sm hover:shadow-md">
                            <Download size={18} /> {t('downloadPdf')}
                        </button>
                    </div>

                    <div id="bill-content" class="bg-white text-slate-800 p-12 rounded-lg shadow-2xl max-w-4xl mx-auto border border-slate-100">
                        {/* Bill Header */}
                        <div class="flex justify-between items-start border-b-2 border-slate-100 pb-8 mb-8">
                            <div>
                                <div class="flex items-center gap-3 mb-4">
                                    <div class="w-12 h-12 bg-indigo-600 rounded-lg flex items-center justify-center">
                                        <span class="text-white font-black text-2xl">R</span>
                                    </div>
                                    <h1 class="text-2xl font-black text-indigo-900 tracking-tight">{t('adminSubtitle')}</h1>
                                </div>
                                <p class="text-slate-500 text-sm max-w-xs">{t('homeCookedSubtitle')}</p>
                            </div>
                            <div class="text-right">
                                <h2 class="text-3xl font-bold text-slate-800 uppercase tracking-tighter">{t('invoice')}</h2>
                                <p class="text-slate-400 mt-1 font-medium">#{report()?.user.user_id}-{new Date().getTime().toString().slice(-6)}</p>
                                <div class="mt-4 text-sm text-center font-bold text-indigo-600 bg-indigo-50 px-4 py-2 rounded-xl whitespace-nowrap">
                                    {new Date(startDate()).toLocaleDateString()} - {new Date(endDate()).toLocaleDateString()}
                                </div>
                            </div>
                        </div>

                        {/* Bill Info */}
                        <div class="grid grid-cols-2 gap-12 mb-12">
                            <div>
                                <h4 class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-3">{t('billTo')}</h4>
                                <p class="text-xl font-bold text-slate-800">{report()?.user.name}</p>
                                <p class="text-slate-500">{report()?.user.mobile_no}</p>
                                <p class="text-slate-500 mt-1">{report()?.user.building_no}, {report()?.user.room_no}</p>
                            </div>
                            <div class="text-right">
                                <h4 class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-3">{t('summary')}</h4>
                                <div class="space-y-1">
                                    <p class="text-slate-500">{t('openingBalance')} <span class="text-slate-800 font-bold">₹{report()?.opening_balance.toFixed(2)}</span></p>
                                    <p class="text-slate-500">{t('totalBillable')} <span class="text-slate-800 font-bold">₹{report()?.total_spent.toFixed(2)}</span></p>
                                    <p class="text-slate-500">{t('recharges')} <span class="text-slate-800 font-bold">₹{report()?.total_recharges.toFixed(2)}</span></p>
                                </div>
                            </div>
                        </div>

                        {/* Bill Table */}
                        <table class="w-full mb-12">
                            <thead>
                                <tr class="bg-slate-50 text-slate-500 text-left">
                                    <th class="py-3 px-4 font-bold text-xs uppercase tracking-wider">{t('date')}</th>
                                    <th class="py-3 px-4 font-bold text-xs uppercase tracking-wider">{t('meal')}</th>
                                    <th class="py-3 px-4 font-bold text-xs uppercase tracking-wider">{t('details')}</th>
                                    <th class="py-3 px-4 font-bold text-xs uppercase tracking-wider text-right">{t('amount')}</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-slate-100">
                                <For each={report()?.logs}>
                                    {(log) => (
                                        <tr class="text-slate-800">
                                            <td class="py-4 px-4 text-sm">{new Date(log.log_date).toLocaleDateString()}</td>
                                            <td class="py-4 px-4 text-sm font-semibold capitalize">{log.meal_type}</td>
                                            <td class="py-4 px-4 text-sm">
                                                {log.has_main_meal ? (
                                                    log.is_special ? (
                                                        <span class="font-bold text-indigo-600">{log.special_dish_name} (S) </span>
                                                    ) : (
                                                        t('standard')
                                                    )
                                                ) : (
                                                    <span class="text-slate-500 italic font-medium">{t('aLaCarte')}</span>
                                                )}
                                                {log.extra_rice_qty > 0 && <span class="text-slate-400">+ {t('rice')} ({log.extra_rice_qty}) </span>}
                                                {log.extra_roti_qty > 0 && <span class="text-slate-400">+ {t('roti')} ({log.extra_roti_qty}) </span>}
                                                {log.extra_chicken_qty > 0 && <span class="text-slate-400">+ {t('chicken')} ({log.extra_chicken_qty}) </span>}
                                                {log.extra_fish_qty > 0 && <span class="text-slate-400">+ {t('fish')} ({log.extra_fish_qty}) </span>}
                                                {log.extra_egg_qty > 0 && <span class="text-slate-400">+ {t('egg')} ({log.extra_egg_qty}) </span>}
                                                {log.extra_vegetable_qty > 0 && <span class="text-slate-400">+ {t('vegetable')} ({log.extra_vegetable_qty}) </span>}
                                            </td>
                                            <td class="py-4 px-4 text-sm font-bold text-right">₹{log.total_cost.toFixed(2)}</td>
                                        </tr>
                                    )}
                                </For>
                            </tbody>
                        </table>

                        {/* Bill Totals */}
                        <div class="flex justify-end pt-8 border-t-2 border-slate-100">
                            <div class="w-64 space-y-4">
                                <div class="flex justify-between text-slate-500 font-medium">
                                    <span>{t('subtotal')}</span>
                                    <span>₹{report()?.total_spent.toFixed(2)}</span>
                                </div>
                                <div class="flex justify-between text-2xl font-black text-indigo-900 border-t border-slate-100 pt-4">
                                    <span>{t('total')}</span>
                                    <span>₹{report()?.total_spent.toFixed(2)}</span>
                                </div>
                                <div class={`p-4 rounded-xl mt-4 flex justify-between ${report()!.closing_balance < 0 ? 'bg-red-50 text-red-700' : 'bg-emerald-50 text-emerald-700'}`}>
                                    <span class="text-xs font-bold uppercase">{t('balanceStatus')}</span>
                                    <span class="font-bold">₹{report()?.closing_balance.toFixed(2)}</span>
                                </div>
                            </div>
                        </div>

                        <div class="mt-20 text-center border-t border-slate-100 pt-8 opacity-40 grayscale">
                            <p class="text-xs font-bold uppercase tracking-[0.2em] mb-2 text-slate-400">{t('generatedVia')}</p>
                            <p class="text-[10px] text-slate-400">{t('thankYou')}</p>
                        </div>
                    </div>
                </div>
            </Show>
        </div>
    );
};

export default Billing;
