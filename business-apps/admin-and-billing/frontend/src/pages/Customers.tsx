import { createSignal, onMount, For } from "solid-js";
import axios from "axios";
import { User } from "../types";
import {
  Plus,
  Search,
  Wallet as WalletIcon,
  Phone,
  MapPin,
  X,
} from "lucide-solid";
import { useI18n } from "../i18n";

import {
  globalUsers,
  globalUserTrie,
  loadUsers,
  appendNewUser,
  updateUserBalance,
} from "../store/userStore";

const Customers = () => {
  const { t } = useI18n();
  const [searchTerm, setSearchTerm] = createSignal("");
  const [showAddModal, setShowAddModal] = createSignal(false);
  const [showRechargeModal, setShowRechargeModal] = createSignal<User | null>(
    null,
  );

  onMount(loadUsers);

  const filteredUsers = () => {
    const term = searchTerm().trim();
    if (!term) return globalUsers();

    const isNumeric = /^[0-9+\s]+$/.test(term);

    if (isNumeric) {
      const stripped = term.replace(/\s+/g, "");
      return globalUsers().filter((u) => u.mobile_no.includes(stripped));
    } else {
      return globalUserTrie().search(term, 100);
    }
  };

  return (
    <div class="space-y-6 animate-in pb-20">
      <header>
        <h2 class="text-2xl font-bold text-slate-800">{t("customers")}</h2>
        <p class="text-slate-500 font-medium text-sm mt-1">
          {t("manageSubscribers")}
        </p>
      </header>

      <button
        onClick={() => setShowAddModal(true)}
        class="btn btn-primary w-full shadow-lg shadow-blue-500/30"
      >
        <Plus size={24} />
        {t("addNewCustomer")}
      </button>

      <div class="relative z-10">
        <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
          <Search class="text-slate-400" size={24} />
        </div>
        <input
          type="text"
          placeholder={t("searchCustomers")}
          class="input-large !pl-14"
          onInput={(e) => setSearchTerm(e.currentTarget.value)}
        />
      </div>

      <div class="space-y-4">
        <For each={filteredUsers()}>
          {(user) => (
            <div class="card p-5 bg-white border border-slate-200">
              <div class="flex justify-between items-start mb-4">
                <div class="flex gap-3">
                  <div class="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center font-bold text-xl text-blue-700">
                    {user.name[0]}
                  </div>
                  <div>
                    <h4 class="font-bold text-xl text-slate-900">
                      {user.name}
                    </h4>
                    <div class="flex items-center gap-2 text-slate-500 text-sm mt-1 font-medium">
                      <Phone size={14} /> {user.mobile_no}
                    </div>
                  </div>
                </div>
                <div class="text-right">
                  <span
                    class={`text-2xl font-black block ${user.balance < 0 ? "text-red-600" : "text-green-600"}`}
                  >
                    ₹{Math.abs(user.balance).toFixed(2)}
                  </span>
                  <span class="text-xs font-bold text-slate-400 uppercase tracking-wider">
                    {user.balance < 0 ? "Due" : "Balance"}
                  </span>
                </div>
              </div>

              <div class="flex gap-2 mb-4 text-sm font-medium">
                <span class="bg-slate-100 text-slate-600 px-3 py-1.5 rounded-lg flex items-center gap-1 flex-1 justify-center">
                  <MapPin size={16} /> Room {user.room_no}
                </span>
                <span class="bg-indigo-50 text-indigo-700 px-3 py-1.5 rounded-lg flex items-center justify-center flex-1">
                  {user.plan === "monthly" ? t("monthly") : t("oneOff")}
                </span>
              </div>

              <button
                onClick={() => setShowRechargeModal(user)}
                class="btn bg-green-50 text-green-700 border-2 border-green-200 hover:bg-green-100 active:bg-green-200 w-full"
              >
                <WalletIcon size={20} />
                {t("recharge")}
              </button>
            </div>
          )}
        </For>
        {filteredUsers().length === 0 && (
          <div class="text-center py-10 text-slate-500 font-medium text-lg">
            No customers found.
          </div>
        )}
      </div>

      {/* Add User Modal */}
      {showAddModal() && (
        <Modal
          title={t("addNewCustomer")}
          onClose={() => setShowAddModal(false)}
        >
          <AddUserForm
            onSuccess={(user) => {
              setShowAddModal(false);
              appendNewUser(user);
            }}
          />
        </Modal>
      )}

      {/* Recharge Modal */}
      {showRechargeModal() && (
        <Modal
          title={showRechargeModal()?.name!}
          onClose={() => setShowRechargeModal(null)}
        >
          <RechargeForm
            user={showRechargeModal()!}
            onSuccess={(newBalance) => {
              updateUserBalance(showRechargeModal()!.user_id, newBalance);
              setShowRechargeModal(null);
            }}
          />
        </Modal>
      )}
    </div>
  );
};

// Bottom Sheet Modal
const Modal = (props: {
  title: string;
  children: any;
  onClose: () => void;
}) => (
  <div class="fixed inset-0 z-[60] bg-slate-900/40 backdrop-blur-sm flex items-end sm:items-center justify-center">
    <div class="bg-white w-full max-w-[600px] rounded-t-3xl sm:rounded-3xl p-6 shadow-2xl pb-safe animate-in slide-in-from-bottom max-h-[90vh] overflow-y-auto">
      <div class="flex justify-between items-center mb-6">
        <h3 class="text-2xl font-bold text-slate-900">{props.title}</h3>
        <button
          onClick={props.onClose}
          class="w-10 h-10 bg-slate-100 rounded-full flex items-center justify-center text-slate-600"
        >
          <X size={24} />
        </button>
      </div>
      {props.children}
    </div>
  </div>
);

const AddUserForm = (props: { onSuccess: (user: User) => void }) => {
  const { t } = useI18n();
  const [formData, setFormData] = createSignal<{
    name: string;
    mobile_no: string;
    building_no: string;
    room_no: string;
    plan: "monthly" | "one_off";
  }>({
    name: "",
    mobile_no: "",
    building_no: "",
    room_no: "",
    plan: "monthly",
  });
  const [isSubmitting, setIsSubmitting] = createSignal(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const res = await axios.post("/api/users", formData());
      props.onSuccess(res.data);
    } catch (err) {
      alert("Failed to add customer");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} class="space-y-4">
      <div>
        <label class="block text-sm font-bold text-slate-700 mb-1">
          {t("fullName")}
        </label>
        <input
          class="input-large !h-14 !text-lg"
          required
          onInput={(e) =>
            setFormData({ ...formData(), name: e.currentTarget.value })
          }
        />
      </div>
      <div>
        <label class="block text-sm font-bold text-slate-700 mb-1">
          {t("mobileNumber")}
        </label>
        <input
          type="tel"
          class="input-large !h-14 !text-lg"
          required
          onInput={(e) =>
            setFormData({ ...formData(), mobile_no: e.currentTarget.value })
          }
        />
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-bold text-slate-700 mb-1">
            {t("buildingNo")}
          </label>
          <input
            class="input-large !h-14 !text-lg"
            required
            onInput={(e) =>
              setFormData({ ...formData(), building_no: e.currentTarget.value })
            }
          />
        </div>
        <div>
          <label class="block text-sm font-bold text-slate-700 mb-1">
            {t("roomNo")}
          </label>
          <input
            class="input-large !h-14 !text-lg"
            required
            onInput={(e) =>
              setFormData({ ...formData(), room_no: e.currentTarget.value })
            }
          />
        </div>
      </div>
      <div>
        <label class="block text-sm font-bold text-slate-700 mb-1">
          {t("planType")}
        </label>
        <div class="bg-slate-200 p-1 rounded-xl flex">
          <button
            type="button"
            onClick={() => setFormData({ ...formData(), plan: "monthly" })}
            class={`flex-1 py-3 text-sm font-bold rounded-lg ${formData().plan === "monthly" ? "bg-white shadow-sm text-indigo-700" : "text-slate-600"}`}
          >
            {t("monthly")}
          </button>
          <button
            type="button"
            onClick={() => setFormData({ ...formData(), plan: "one_off" })}
            class={`flex-1 py-3 text-sm font-bold rounded-lg ${formData().plan === "one_off" ? "bg-white shadow-sm text-indigo-700" : "text-slate-600"}`}
          >
            {t("oneOff")}
          </button>
        </div>
      </div>
      <button
        type="submit"
        disabled={isSubmitting()}
        class="btn btn-primary w-full mt-6"
      >
        {isSubmitting() ? "Saving..." : t("saveCustomer")}
      </button>
    </form>
  );
};

const RechargeForm = (props: {
  user: User;
  onSuccess: (newBalance: number) => void;
}) => {
  const { t } = useI18n();
  const [amount, setAmount] = createSignal("");
  const [refId, setRefId] = createSignal("");
  const [isSubmitting, setIsSubmitting] = createSignal(false);

  const getCurrentDateTime = () => {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, "0");
    const day = String(now.getDate()).padStart(2, "0");
    const hours = String(now.getHours()).padStart(2, "0");
    const minutes = String(now.getMinutes()).padStart(2, "0");
    return `${year}-${month}-${day}T${hours}:${minutes}`;
  };

  const [txnDateTime, setTxnDateTime] = createSignal(getCurrentDateTime());

  const quickAmounts = [100, 500, 1000];

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      let dtString = txnDateTime();
      if (dtString.length === 16) {
        dtString += ":59.000";
      }
      const timestamp = new Date(dtString).toISOString();

      const res = await axios.post("/api/wallet/recharge", {
        user_id: props.user.user_id,
        amount: parseFloat(amount()),
        ref_id: refId(),
        txn_date: timestamp,
      });
      props.onSuccess(res.data.new_balance);
    } catch (err) {
      alert("Failed to recharge");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} class="space-y-6">
      <div>
        <label class="block text-sm font-bold text-slate-700 mb-2">
          {t("amountReq")}
        </label>
        <div class="relative">
          <span class="absolute left-4 top-1/2 -translate-y-1/2 text-2xl font-bold text-slate-400">
            ₹
          </span>
          <input
            type="text"
            inputMode="decimal"
            pattern="[0-9]*[.,]?[0-9]*"
            class="input-large !text-3xl !font-black !pl-12"
            required
            value={amount()}
            onInput={(e) => {
              // allow users to type comma as decimal separator and normalize to dot
              const raw = e.currentTarget.value;
              const normalized = raw.replace(/,/g, ".");
              // allow empty string, numbers, and partial decimals while typing
              if (/^\d*(?:\.\d{0,2})?$/.test(normalized) || normalized === "") {
                setAmount(normalized);
              } else {
                // if user types other chars, strip them
                const cleaned = normalized.replace(/[^0-9.]/g, "");
                setAmount(cleaned);
              }
            }}
          />
        </div>

        <div class="flex gap-2 mt-3">
          <For each={quickAmounts}>
            {(amt) => (
              <button
                type="button"
                onClick={() => setAmount(amt.toString())}
                class="flex-1 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold rounded-lg border border-slate-200"
              >
                +₹{amt}
              </button>
            )}
          </For>
        </div>
      </div>

      <div>
        <label class="block text-sm font-bold text-slate-700 mb-1">
          {t("txnDateTime")}
        </label>
        <input
          type="datetime-local"
          class="input-large !h-14 !text-lg"
          required
          value={txnDateTime()}
          onInput={(e) => setTxnDateTime(e.currentTarget.value)}
        />
      </div>
      <div>
        <label class="block text-sm font-bold text-slate-700 mb-1">
          {t("paymentRef")}
        </label>
        <input
          class="input-large !h-14 !text-lg"
          placeholder={t("optional")}
          value={refId()}
          onInput={(e) => setRefId(e.currentTarget.value)}
        />
      </div>
      <button
        type="submit"
        disabled={isSubmitting()}
        class="btn btn-success w-full mt-2 text-xl"
      >
        {isSubmitting() ? "Processing..." : t("confirmRecharge")}
      </button>
    </form>
  );
};

export default Customers;
