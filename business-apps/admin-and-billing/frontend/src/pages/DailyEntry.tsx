import axios from 'axios';
import { Check, ChefHat, Moon, Salad, SquarePen, Sun, Trash2, Utensils, X, Calendar, Search } from 'lucide-solid';
import { For, createEffect, createSignal, onMount } from 'solid-js';
import { Portal } from 'solid-js/web';
import { DailyLog, User } from '../types';
import { useI18n } from '../i18n';

import { globalUsers, globalUserTrie, loadUsers, updateUserBalance } from '../store/userStore';

const DailyEntry = () => {
    const { t } = useI18n();
    const [prices, setPrices] = createSignal<Record<string, number>>({});
    const [logs, setLogs] = createSignal<DailyLog[]>([]);
    const [selectedUser, setSelectedUser] = createSignal<string>('');
    const [searchQuery, setSearchQuery] = createSignal('');
    const [suggestions, setSuggestions] = createSignal<User[]>([]);
    const [showSuggestions, setShowSuggestions] = createSignal(false);
    
    // Better date handling for UI
    const getToday = () => new Date().toISOString().split('T')[0];
    const getYesterday = () => {
        const d = new Date();
        d.setDate(d.getDate() - 1);
        return d.toISOString().split('T')[0];
    };
    const [date, setDate] = createSignal(getToday());
    
    const [mealType, setMealType] = createSignal<'lunch' | 'dinner'>('lunch');
    const [mealCategory, setMealCategory] = createSignal<'standard' | 'special' | 'none'>('standard');
    const [specialDish, setSpecialDish] = createSignal('');
    const [extraRice, setExtraRice] = createSignal(0);
    const [extraRoti, setExtraRoti] = createSignal(0);
    const [extraChicken, setExtraChicken] = createSignal(0);
    const [extraFish, setExtraFish] = createSignal(0);
    const [extraEgg, setExtraEgg] = createSignal(0);
    const [extraVegetable, setExtraVegetable] = createSignal(0);
    const [isSubmitting, setIsSubmitting] = createSignal(false);
    const [successMsg, setSuccessMsg] = createSignal(false);
    const [editingLog, setEditingLog] = createSignal<DailyLog | null>(null);

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

    const fetchPrices = async () => {
        try {
            const res = await axios.get('/api/meals');
            const priceMap: Record<string, number> = {};
            res.data?.forEach((item: any) => {
                priceMap[item.item_id] = item.price;
            });
            setPrices(priceMap);
        } catch (error) {
            console.error('Failed to fetch prices:', error);
        }
    };

    onMount(() => {
        loadUsers();
        fetchPrices();
        
        // Auto-select shift based on time
        const hour = new Date().getHours();
        if (hour >= 16) {
            setMealType('dinner');
        }
    });

    const fetchLogs = async () => {
        try {
            const userId = selectedUser();
            const res = await axios.get(`/api/daily-entry?date=${date()}${userId ? `&user_id=${userId}` : ''}`);
            setLogs(res.data || []);
        } catch (error) {
            console.error('Failed to fetch logs:', error);
        }
    };

    createEffect(() => {
        date();
        selectedUser();
        fetchLogs();
    });

    const handleDelete = async (logId: number) => {
        const logToDelete = logs().find(l => l.log_id === logId);
        if (!logToDelete) return;
        if (!confirm('Are you sure you want to delete this entry? This will refund the cost.')) return;

        try {
            const res = await axios.delete(`/api/daily-entry/${logId}`);
            setSuccessMsg(true);
            setTimeout(() => setSuccessMsg(false), 3000);
            await fetchLogs();
            updateUserBalance(logToDelete.user_id, res.data.new_balance);
        } catch (error) {
            console.error('Failed to delete log:', error);
            alert('Failed to delete entry');
        }
    };

    const handleSubmit = async (e?: Event) => {
        if (e) e.preventDefault();
        if (!selectedUser() || !date()) {
            alert('Please select a customer.');
            return;
        }

        setIsSubmitting(true);
        try {
            const res = await axios.post('/api/daily-entry', {
                user_id: parseInt(selectedUser()),
                log_date: new Date(date()).toISOString(),
                meal_type: mealType(),
                has_main_meal: mealCategory() !== 'none',
                is_special: mealCategory() === 'special',
                special_dish_name: mealCategory() === 'special' ? specialDish() : '',
                extra_rice_qty: extraRice(),
                extra_roti_qty: extraRoti(),
                extra_chicken_qty: extraChicken(),
                extra_fish_qty: extraFish(),
                extra_egg_qty: extraEgg(),
                extra_vegetable_qty: extraVegetable()
            });
            setSuccessMsg(true);
            setTimeout(() => setSuccessMsg(false), 3000);
            updateUserBalance(parseInt(selectedUser()), res.data.new_balance);
            await fetchLogs();

            // Reset fields
            setMealCategory('standard');
            setSpecialDish('');
            setExtraRice(0);
            setExtraRoti(0);
            setExtraChicken(0);
            setExtraFish(0);
            setExtraEgg(0);
            setExtraVegetable(0);
            setSearchQuery('');
            setSelectedUser('');
        } catch (err) {
            alert('Failed to record entry');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div class="space-y-6 animate-in">
            <header>
                <h2 class="text-2xl font-bold text-slate-800">{t('dailyEntry')}</h2>
                <p class="text-slate-500 font-medium text-sm mt-1">{t('recordConsumption')}</p>
            </header>

            <form onSubmit={handleSubmit} class="space-y-4">
                
                {/* Search Customer (Huge input) */}
                <div class="relative z-40">
                    <div class="relative">
                        <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                            <Search class="text-slate-400" size={24} />
                        </div>
                        <input
                            type="text"
                            class={`input-large !pl-14 ${!selectedUser() && searchQuery().length > 0 ? 'border-red-400' : 'border-blue-400'}`}
                            placeholder={t('selectCustomer') || "Search Customer Name..."}
                            value={searchQuery()}
                            onInput={handleSearchInput}
                            onFocus={() => { if (searchQuery().trim().length >= 3) setShowSuggestions(true); }}
                            onBlur={handleSearchBlur}
                            required={!selectedUser()}
                        />
                    </div>
                    {showSuggestions() && (
                        <ul class="absolute top-full left-0 right-0 mt-2 bg-white rounded-xl shadow-xl border border-slate-200 overflow-hidden">
                            <For each={suggestions()}>
                                {(user) => (
                                    <li 
                                        class="p-4 cursor-pointer hover:bg-slate-50 border-b border-slate-100 last:border-0 flex justify-between items-center transition-colors active:bg-blue-50"
                                        onClick={() => handleSelectUser(user)}
                                    >
                                        <div>
                                            <span class="font-bold text-lg text-slate-900 block">{user.name}</span>
                                            <span class="text-sm text-slate-500">Room {user.room_no}</span>
                                        </div>
                                        <span class={`font-bold text-lg ${user.balance < 0 ? 'text-red-600' : 'text-green-600'}`}>
                                            ₹{Math.abs(user.balance).toFixed(0)}
                                        </span>
                                    </li>
                                )}
                            </For>
                            {suggestions().length === 0 && (
                                <li class="p-6 text-center text-slate-500 font-medium text-lg">
                                    No customers found
                                </li>
                            )}
                        </ul>
                    )}
                </div>

                {/* Date & Shift Segmented Controls */}
                <div class="grid grid-cols-2 gap-4">
                    <div class="bg-slate-200 p-1 rounded-xl flex">
                        <button
                            type="button"
                            onClick={() => setDate(getToday())}
                            class={`flex-1 py-3 text-sm font-bold rounded-lg transition-colors ${date() === getToday() ? 'bg-white shadow-sm text-blue-700' : 'text-slate-600'}`}
                        >
                            Today
                        </button>
                        <button
                            type="button"
                            onClick={() => setDate(getYesterday())}
                            class={`flex-1 py-3 text-sm font-bold rounded-lg transition-colors ${date() === getYesterday() ? 'bg-white shadow-sm text-blue-700' : 'text-slate-600'}`}
                        >
                            Yesterday
                        </button>
                    </div>

                    <div class="bg-slate-200 p-1 rounded-xl flex">
                        <button
                            type="button"
                            onClick={() => setMealType('lunch')}
                            class={`flex-1 py-3 text-sm font-bold rounded-lg transition-colors flex items-center justify-center gap-1 ${mealType() === 'lunch' ? 'bg-white shadow-sm text-amber-600' : 'text-slate-600'}`}
                        >
                            <Sun size={16} /> {t('lunch')}
                        </button>
                        <button
                            type="button"
                            onClick={() => setMealType('dinner')}
                            class={`flex-1 py-3 text-sm font-bold rounded-lg transition-colors flex items-center justify-center gap-1 ${mealType() === 'dinner' ? 'bg-white shadow-sm text-indigo-600' : 'text-slate-600'}`}
                        >
                            <Moon size={16} /> {t('dinner')}
                        </button>
                    </div>
                </div>

                {/* Main Meal Cards */}
                <div class="grid grid-cols-3 gap-3 pt-2">
                    <label class={`card p-3 flex flex-col items-center justify-center gap-2 cursor-pointer transition-all border-2 ${mealCategory() === 'standard' ? 'border-blue-500 bg-blue-50' : 'border-transparent'}`}>
                        <input type="radio" name="cat" class="hidden" checked={mealCategory() === 'standard'} onChange={() => setMealCategory('standard')} />
                        <Utensils size={32} class={mealCategory() === 'standard' ? 'text-blue-600' : 'text-slate-400'} />
                        <span class={`font-bold text-sm ${mealCategory() === 'standard' ? 'text-blue-900' : 'text-slate-600'}`}>{t('standard')}</span>
                    </label>
                    <label class={`card p-3 flex flex-col items-center justify-center gap-2 cursor-pointer transition-all border-2 ${mealCategory() === 'special' ? 'border-purple-500 bg-purple-50' : 'border-transparent'}`}>
                        <input type="radio" name="cat" class="hidden" checked={mealCategory() === 'special'} onChange={() => setMealCategory('special')} />
                        <ChefHat size={32} class={mealCategory() === 'special' ? 'text-purple-600' : 'text-slate-400'} />
                        <span class={`font-bold text-sm ${mealCategory() === 'special' ? 'text-purple-900' : 'text-slate-600'}`}>Special</span>
                    </label>
                    <label class={`card p-3 flex flex-col items-center justify-center gap-2 cursor-pointer transition-all border-2 ${mealCategory() === 'none' ? 'border-orange-500 bg-orange-50' : 'border-transparent'}`}>
                        <input type="radio" name="cat" class="hidden" checked={mealCategory() === 'none'} onChange={() => setMealCategory('none')} />
                        <Salad size={32} class={mealCategory() === 'none' ? 'text-orange-600' : 'text-slate-400'} />
                        <span class={`font-bold text-sm text-center ${mealCategory() === 'none' ? 'text-orange-900' : 'text-slate-600'}`}>{t('extrasOnly')}</span>
                    </label>
                </div>

                {/* Special Dish Input */}
                {mealCategory() === 'special' && (
                    <div class="animate-in slide-in-from-top-2">
                        <input
                            class="input-large !border-purple-300 focus:!border-purple-500 !bg-purple-50"
                            placeholder={t('dishName') + ' (e.g., Mutton)'}
                            required
                            value={specialDish()}
                            onInput={e => setSpecialDish(e.currentTarget.value)}
                        />
                    </div>
                )}

                {/* Big Extras Controls */}
                <div class="pt-2 pb-24">
                    <h4 class="font-bold text-slate-800 mb-3">{t('extras')}</h4>
                    <div class="grid grid-cols-2 gap-3">
                        <ExtraItem label={t('rice')} price={prices()['rice'] ?? 10} value={extraRice()} onChange={setExtraRice} />
                        <ExtraItem label={t('roti')} price={prices()['roti'] ?? 4} value={extraRoti()} onChange={setExtraRoti} />
                        <ExtraItem label={t('chicken')} price={prices()['chicken'] ?? 30} value={extraChicken()} onChange={setExtraChicken} />
                        <ExtraItem label={t('fish')} price={prices()['fish'] ?? 20} value={extraFish()} onChange={setExtraFish} />
                        <ExtraItem label={t('egg')} price={prices()['egg'] ?? 10} value={extraEgg()} onChange={setExtraEgg} />
                        <ExtraItem label={t('vegetable')} price={prices()['vegetable'] ?? 15} value={extraVegetable()} onChange={setExtraVegetable} />
                    </div>
                </div>

                {/* Submit button moved to Portal */}
            </form>

            {/* Toast moved to Portal */}

            {/* Recent Entries */}
            <div class="pt-6 pb-20">
                <h3 class="font-bold text-slate-800 mb-4">{t('entries')} for {date()}</h3>
                
                <div class="space-y-3">
                    <For each={logs()}>
                        {(log) => (
                            <div class="card p-4 relative overflow-hidden group">
                                <div class={`absolute top-0 bottom-0 left-0 w-2 ${log.meal_type === 'lunch' ? 'bg-amber-400' : 'bg-indigo-500'}`}></div>
                                
                                <div class="pl-2 flex justify-between items-start mb-2">
                                    <div>
                                        <p class="font-bold text-lg text-slate-900">
                                            {log.user_name || globalUsers().find(u => u.user_id === log.user_id)?.name || log.user_id}
                                        </p>
                                        <p class="text-sm text-slate-500">
                                            {log.has_main_meal ? (log.is_special ? log.special_dish_name : t('standard')) : t('extrasOnly')}
                                        </p>
                                    </div>
                                    <p class="font-black text-xl text-slate-800">₹{log.total_cost}</p>
                                </div>
                                
                                <div class="pl-2 flex justify-between items-end">
                                    <div class="text-xs text-slate-500 flex flex-wrap gap-1 max-w-[70%]">
                                        {log.extra_rice_qty > 0 && <span class="bg-slate-100 px-2 py-1 rounded">Rice: {log.extra_rice_qty}</span>}
                                        {log.extra_roti_qty > 0 && <span class="bg-slate-100 px-2 py-1 rounded">Roti: {log.extra_roti_qty}</span>}
                                        {log.extra_chicken_qty > 0 && <span class="bg-slate-100 px-2 py-1 rounded">Chicken: {log.extra_chicken_qty}</span>}
                                        {log.extra_fish_qty > 0 && <span class="bg-slate-100 px-2 py-1 rounded">Fish: {log.extra_fish_qty}</span>}
                                        {log.extra_egg_qty > 0 && <span class="bg-slate-100 px-2 py-1 rounded">Egg: {log.extra_egg_qty}</span>}
                                        {log.extra_vegetable_qty > 0 && <span class="bg-slate-100 px-2 py-1 rounded">Veg: {log.extra_vegetable_qty}</span>}
                                    </div>
                                    
                                    <div class="flex gap-2">
                                        <button
                                            class="w-10 h-10 rounded-full bg-slate-100 text-slate-600 flex items-center justify-center hover:bg-slate-200 active:bg-slate-300"
                                            onClick={() => setEditingLog(log)}
                                        >
                                            <SquarePen size={18} />
                                        </button>
                                        <button
                                            class="w-10 h-10 rounded-full bg-red-100 text-red-600 flex items-center justify-center hover:bg-red-200 active:bg-red-300"
                                            onClick={() => handleDelete(log.log_id)}
                                        >
                                            <Trash2 size={18} />
                                        </button>
                                    </div>
                                </div>
                            </div>
                        )}
                    </For>
                    {logs().length === 0 && (
                        <div class="card p-8 text-center text-slate-500 font-medium">
                            {t('noEntries')}
                        </div>
                    )}
                </div>
            </div>

            <Portal>
                {/* Fixed Submit Button */}
                <div class="fixed bottom-[72px] left-0 right-0 max-w-[600px] mx-auto p-4 bg-white border-t border-slate-200 z-30 shadow-[0_-10px_20px_rgba(0,0,0,0.05)] pointer-events-auto">
                    <button
                        type="button"
                        onClick={() => handleSubmit()}
                        disabled={isSubmitting()}
                        class={`btn btn-primary w-full shadow-lg shadow-blue-500/30 ${isSubmitting() ? 'opacity-50' : ''}`}
                    >
                        <Check size={24} />
                        {isSubmitting() ? t('recording') : t('recordEntry')}
                    </button>
                </div>

                {/* Success Message Toast */}
                <div class={`fixed top-20 left-1/2 -translate-x-1/2 z-50 bg-green-600 text-white px-6 py-3 rounded-full shadow-xl flex items-center gap-2 transition-all duration-300 pointer-events-none ${successMsg() ? 'translate-y-0 opacity-100' : '-translate-y-20 opacity-0'}`}>
                    <Check size={20} />
                    <span class="font-bold">{t('mealRecorded')}</span>
                </div>

                {/* Edit Modal */}
                {editingLog() && (
                    <div class="fixed inset-0 z-[60] bg-slate-900/40 backdrop-blur-sm flex items-end justify-center sm:items-center pointer-events-auto">
                        <div class="bg-white w-full max-w-[600px] rounded-t-3xl sm:rounded-3xl p-6 shadow-2xl pb-safe animate-in slide-in-from-bottom">
                            <div class="flex justify-between items-center mb-6">
                                <h3 class="text-xl font-bold text-slate-900">Edit Extras</h3>
                                <button onClick={() => setEditingLog(null)} class="w-10 h-10 bg-slate-100 rounded-full flex items-center justify-center text-slate-600">
                                    <X size={24} />
                                </button>
                            </div>
                            
                            <EditLogContent 
                                log={editingLog()!} 
                                onSuccess={(bal) => {
                                    setEditingLog(null);
                                    fetchLogs();
                                    updateUserBalance(editingLog()!.user_id, bal);
                                    setSuccessMsg(true);
                                    setTimeout(() => setSuccessMsg(false), 3000);
                                }}
                            />
                        </div>
                    </div>
                )}
            </Portal>
        </div>
    );
};

// Extracted Edit content to manage its own state properly
const EditLogContent = (props: { log: DailyLog; onSuccess: (newBalance: number) => void }) => {
    const [mealType, setMealType] = createSignal(props.log.meal_type);
    const [extraRice, setExtraRice] = createSignal(props.log.extra_rice_qty);
    const [extraRoti, setExtraRoti] = createSignal(props.log.extra_roti_qty);
    const [isSubmitting, setIsSubmitting] = createSignal(false);

    const handleSubmit = async () => {
        setIsSubmitting(true);
        try {
            const res = await axios.put(`/api/daily-entry/${props.log.log_id}`, {
                meal_type: mealType(),
                extra_rice_qty: extraRice(),
                extra_roti_qty: extraRoti(),
                has_main_meal: props.log.has_main_meal,
                is_special: props.log.is_special
            });
            props.onSuccess(res.data.new_balance);
        } catch (err) {
            alert('Failed to update');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div class="space-y-6">
            <div class="bg-slate-200 p-1 rounded-xl flex">
                <button type="button" onClick={() => setMealType('lunch')} class={`flex-1 py-3 text-sm font-bold rounded-lg ${mealType() === 'lunch' ? 'bg-white shadow-sm text-amber-600' : 'text-slate-600'}`}>Lunch</button>
                <button type="button" onClick={() => setMealType('dinner')} class={`flex-1 py-3 text-sm font-bold rounded-lg ${mealType() === 'dinner' ? 'bg-white shadow-sm text-indigo-600' : 'text-slate-600'}`}>Dinner</button>
            </div>
            
            <div class="grid grid-cols-2 gap-3">
                <ExtraItem label="Rice" price={0} value={extraRice()} onChange={setExtraRice} />
                <ExtraItem label="Roti" price={0} value={extraRoti()} onChange={setExtraRoti} />
            </div>

            <button onClick={handleSubmit} disabled={isSubmitting()} class="btn btn-primary w-full h-14">
                {isSubmitting() ? 'Saving...' : 'Save Changes'}
            </button>
        </div>
    );
}

const ExtraItem = (props: { label: string, price: number, value: number, onChange: (v: number) => void }) => (
    <div class="card p-3 flex flex-col items-center bg-white shadow-sm border border-slate-200">
        <span class="font-bold text-slate-800">{props.label}</span>
        {props.price > 0 && <span class="text-xs text-slate-400 mb-2">₹{props.price}</span>}
        <div class="flex items-center gap-3 mt-1 w-full justify-center">
            <button type="button" onClick={() => props.onChange(Math.max(0, props.value - 1))} class="w-12 h-12 rounded-full bg-slate-100 active:bg-slate-300 text-slate-700 font-bold text-2xl flex items-center justify-center">-</button>
            <span class="text-2xl font-black text-slate-900 w-6 text-center">{props.value}</span>
            <button type="button" onClick={() => props.onChange(props.value + 1)} class="w-12 h-12 rounded-full bg-blue-100 active:bg-blue-300 text-blue-700 font-bold text-2xl flex items-center justify-center">+</button>
        </div>
    </div>
);

export default DailyEntry;
