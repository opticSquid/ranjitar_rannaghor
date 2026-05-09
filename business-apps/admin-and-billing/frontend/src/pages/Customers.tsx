import { createSignal, onMount, For } from 'solid-js';
import axios from 'axios';
import { User } from '../types';
import { Plus, Search, Wallet as WalletIcon } from 'lucide-solid';
import { useI18n } from '../i18n';

import { globalUsers, globalUserTrie, loadUsers, appendNewUser, updateUserBalance } from '../store/userStore';

const Customers = () => {
    const { t } = useI18n();
    const [searchTerm, setSearchTerm] = createSignal('');
    const [showAddModal, setShowAddModal] = createSignal(false);
    const [showRechargeModal, setShowRechargeModal] = createSignal<User | null>(null);

    onMount(loadUsers);

    const filteredUsers = () => {
        const term = searchTerm().trim();
        if (!term) return globalUsers();

        const isNumeric = /^[0-9+\s]+$/.test(term);

        if (isNumeric) {
            const stripped = term.replace(/\s+/g, '');
            return globalUsers().filter(u => u.mobile_no.includes(stripped));
        } else {
            return globalUserTrie().search(term, 1000); // Higher limit for grid view
        }
    };

    return (
        <div class="space-y-8 animate-in slide-in-from-bottom duration-500">
            <header class="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h2 class="text-3xl font-bold text-white">{t('customers')}</h2>
                    <p class="text-text-dim mt-2">{t('manageSubscribers')}</p>
                </div>
                <button
                    onClick={() => setShowAddModal(true)}
                    class="btn btn-primary"
                >
                    <Plus size={20} /> {t('addNewCustomer')}
                </button>
            </header>

            <div class="flex items-center gap-4 glass p-4 bg-white/5 border-none">
                <Search size={20} class="text-text-dim ml-2" />
                <input
                    type="text"
                    placeholder={t('searchCustomers')}
                    class="bg-transparent border-none outline-none text-white w-full placeholder:text-text-dim"
                    onInput={(e) => setSearchTerm(e.currentTarget.value)}
                />
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                <For each={filteredUsers()}>
                    {(user) => (
                        <div class="glass p-6 card border-white/5 hover:border-primary/50 group">
                            <div class="flex items-center gap-4 mb-4">
                                <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-primary to-indigo-600 flex items-center justify-center font-bold text-lg shadow-lg">
                                    {user.name[0]}
                                </div>
                                <div>
                                    <h4 class="font-bold text-lg group-hover:text-primary transition-colors">{user.name}</h4>
                                    <p class="text-text-dim text-sm">{user.mobile_no}</p>
                                </div>
                            </div>

                            <div class="space-y-3 pt-4 border-t border-white/5">
                                <div class="flex justify-between text-sm">
                                    <span class="text-text-dim">{t('location')}</span>
                                    <span>{user.building_no}, {user.room_no}</span>
                                </div>
                                <div class="flex justify-between text-sm">
                                    <span class="text-text-dim">{t('plan')}</span>
                                    <span class="capitalize text-accent font-medium">{user.plan === 'monthly' ? t('monthly') : t('oneOff')}</span>
                                </div>
                                <div class="flex justify-between items-center bg-white/5 p-3 rounded-xl mt-4">
                                    <span class="text-xs font-semibold uppercase tracking-wider text-text-dim">{t('wallet')}</span>
                                    <span class={`text-lg font-bold ${user.balance < 0 ? 'text-error' : 'text-success'}`}>
                                        ₹{user.balance.toFixed(2)}
                                    </span>
                                </div>
                            </div>

                            <div class="mt-6 flex gap-2">
                                <button
                                    onClick={() => setShowRechargeModal(user)}
                                    class="flex-1 btn bg-white/5 hover:bg-white/10 text-white border border-white/10 flex items-center justify-center gap-2"
                                >
                                    <WalletIcon size={18} /> {t('recharge')}
                                </button>
                            </div>
                        </div>
                    )}
                </For>
            </div>

            {/* Add User Modal */}
            {showAddModal() && (
                <Modal title={t('addNewCustomer')} onClose={() => setShowAddModal(false)}>
                    <AddUserForm onSuccess={(user) => { setShowAddModal(false); appendNewUser(user); }} />
                </Modal>
            )}

            {/* Recharge Modal */}
            {showRechargeModal() && (
                <Modal title={`${t('rechargeWallet')}: ${showRechargeModal()?.name}`} onClose={() => setShowRechargeModal(null)}>
                    <RechargeForm user={showRechargeModal()!} onSuccess={(newBalance) => {
                        updateUserBalance(showRechargeModal()!.user_id, newBalance);
                        setShowRechargeModal(null);
                    }} />
                </Modal>
            )}
        </div>
    );
};

const Modal = (props: { title: string; children: any; onClose: () => void }) => (
    <div class="fixed inset-0 z-50 grid place-items-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-300">
        <div class="glass w-full max-w-md p-8 shadow-2xl animate-in zoom-in duration-300">
            <div class="flex justify-between items-center mb-6">
                <h3 class="text-2xl font-bold">{props.title}</h3>
                <button onClick={props.onClose} class="text-text-dim hover:text-white">&times;</button>
            </div>
            {props.children}
        </div>
    </div>
);

const AddUserForm = (props: { onSuccess: (user: User) => void }) => {
    const { t } = useI18n();
    const [formData, setFormData] = createSignal({
        name: '',
        mobile_no: '',
        building_no: '',
        room_no: '',
        plan: 'monthly' as const
    });

    const handleSubmit = async (e: Event) => {
        e.preventDefault();
        try {
            const res = await axios.post('/api/users', formData());
            props.onSuccess(res.data);
        } catch (err) {
            alert('Failed to add customer');
        }
    };

    return (
        <form onSubmit={handleSubmit} class="space-y-4">
            <div>
                <label class="block text-sm font-medium text-text-dim mb-1">{t('fullName')}</label>
                <input
                    class="input"
                    required
                    onInput={e => setFormData({ ...formData(), name: e.currentTarget.value })}
                />
            </div>
            <div>
                <label class="block text-sm font-medium text-text-dim mb-1">{t('mobileNumber')}</label>
                <input
                    class="input"
                    required
                    onInput={e => setFormData({ ...formData(), mobile_no: e.currentTarget.value })}
                />
            </div>
            <div class="grid grid-cols-2 gap-4">
                <div>
                    <label class="block text-sm font-medium text-text-dim mb-1">{t('buildingNo')}</label>
                    <input
                        class="input"
                        required
                        onInput={e => setFormData({ ...formData(), building_no: e.currentTarget.value })}
                    />
                </div>
                <div>
                    <label class="block text-sm font-medium text-text-dim mb-1">{t('roomNo')}</label>
                    <input
                        class="input"
                        required
                        onInput={e => setFormData({ ...formData(), room_no: e.currentTarget.value })}
                    />
                </div>
            </div>
            <div>
                <label class="block text-sm font-medium text-text-dim mb-1">{t('planType')}</label>
                <select
                    class="input bg-surface"
                    onInput={e => setFormData({ ...formData(), plan: e.currentTarget.value as any })}
                >
                    <option value="monthly">{t('monthly')}</option>
                    <option value="one_off">{t('oneOff')}</option>
                </select>
            </div>
            <button type="submit" class="btn btn-primary w-full mt-4">{t('saveCustomer')}</button>
        </form>
    );
};

const RechargeForm = (props: { user: User; onSuccess: (newBalance: number) => void }) => {
    const { t } = useI18n();
    const [amount, setAmount] = createSignal('');
    const [refId, setRefId] = createSignal('');

    // Get current datetime in local format for datetime-local input
    const getCurrentDateTime = () => {
        const now = new Date();
        const year = now.getFullYear();
        const month = String(now.getMonth() + 1).padStart(2, '0');
        const day = String(now.getDate()).padStart(2, '0');
        const hours = String(now.getHours()).padStart(2, '0');
        const minutes = String(now.getMinutes()).padStart(2, '0');
        return `${year}-${month}-${day}T${hours}:${minutes}`;
    };

    const [txnDateTime, setTxnDateTime] = createSignal(getCurrentDateTime());

    const handleSubmit = async (e: Event) => {
        e.preventDefault();
        try {
            // adding seconds part to timestamp
            setTxnDateTime(txnDateTime() + ":59.000000")
            // Convert datetime-local value to ISO timestamp
            const timestamp = new Date(txnDateTime()).toISOString();

            const res = await axios.post('/api/wallet/recharge', {
                user_id: props.user.user_id,
                amount: parseFloat(amount()),
                ref_id: refId(),
                txn_date: timestamp
            });
            props.onSuccess(res.data.new_balance);
        } catch (err) {
            alert('Failed to recharge');
        }
    };

    return (
        <form onSubmit={handleSubmit} class="space-y-4">
            <div>
                <label class="block text-sm font-medium text-text-dim mb-1">{t('amountReq')}</label>
                <input
                    type="text"
                    inputMode='decimal'
                    class="input text-2xl font-bold"
                    required
                    onInput={e => setAmount(e.currentTarget.value)}
                />
            </div>
            <div>
                <label class="block text-sm font-medium text-text-dim mb-1">{t('txnDateTime')}</label>
                <input
                    type="datetime-local"
                    class="input"
                    required
                    value={txnDateTime()}
                    onInput={e => setTxnDateTime(e.currentTarget.value)}
                />
            </div>
            <div>
                <label class="block text-sm font-medium text-text-dim mb-1">{t('paymentRef')}</label>
                <input
                    class="input"
                    placeholder={t('optional')}
                    value={refId()}
                    onInput={e => setRefId(e.currentTarget.value)}
                />
            </div>
            <button type="submit" class="btn btn-primary w-full mt-4">{t('confirmRecharge')}</button>
        </form>
    );
};

export default Customers;
