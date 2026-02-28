"use client";

import React, { useState } from "react";

export default function WalletPage() {
  const [balance] = useState(125.50);
  const [transactions] = useState([
    { id: "tx_1", title: "Order #M1029", type: "debit", amount: 12.50, date: "2024-10-20T14:30:00Z" },
    { id: "tx_2", title: "Cashback from Promo", type: "credit", amount: 5.00, date: "2024-10-18T10:15:00Z" },
    { id: "tx_3", title: "Top-up via bKash", type: "credit", amount: 50.00, date: "2024-10-15T09:00:00Z" },
    { id: "tx_4", title: "Order #M1015", type: "debit", amount: 25.00, date: "2024-10-12T19:45:00Z" },
  ]);

  return (
    <div className="bg-white rounded-3xl p-6 sm:p-8 shadow-sm border border-neutral-100">
      <div className="flex items-center justify-between mb-8">
        <h2 className="text-2xl font-bold text-neutral-900">Wallet</h2>
      </div>

      {/* Balance Card */}
      <div className="bg-gradient-to-tr from-neutral-900 to-neutral-800 p-8 rounded-3xl text-white shadow-xl shadow-neutral-200 mb-10 overflow-hidden relative">
        <div className="absolute top-0 right-0 w-64 h-64 bg-white/10 rounded-full blur-3xl -mr-20 -mt-20 pointer-events-none" />
        <div className="relative z-10 flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
           <div>
             <p className="text-sm font-bold text-neutral-400 uppercase tracking-widest mb-1">Available Balance</p>
             <h3 className="text-5xl font-black">${balance.toFixed(2)}</h3>
           </div>
           
           <div className="flex flex-col sm:flex-row w-full md:w-auto gap-3">
              <button className="bg-orange-500 hover:bg-orange-600 text-white font-extrabold py-3 px-8 rounded-2xl transition-all shadow-lg active:scale-[0.98] whitespace-nowrap">
                Top Up
              </button>
              <button className="bg-white/10 hover:bg-white/20 text-white border border-white/20 font-extrabold py-3 px-8 rounded-2xl transition-all active:scale-[0.98] whitespace-nowrap">
                Withdraw
              </button>
           </div>
        </div>
      </div>

      <h3 className="text-lg font-bold text-neutral-900 mb-6">Recent Transactions</h3>
      
      {transactions.length === 0 ? (
         <div className="text-center py-12 border border-neutral-100 rounded-2xl bg-neutral-50/50">
            <div className="w-16 h-16 bg-white rounded-full flex mx-auto items-center justify-center text-3xl mb-4 shadow-sm border border-neutral-100">📄</div>
            <p className="text-neutral-500 font-medium">No transactions yet.</p>
         </div>
      ) : (
        <div className="space-y-4">
          {transactions.map(tx => (
            <div key={tx.id} className="flex items-center justify-between p-4 rounded-2xl hover:bg-neutral-50 border border-transparent hover:border-neutral-100 transition-colors">
               <div className="flex items-center gap-4">
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center text-xl shrink-0 ${
                    tx.type === "credit" ? "bg-green-100 text-green-600" : "bg-red-100 text-red-600"
                  }`}>
                    {tx.type === "credit" ? "↓" : "↑"}
                  </div>
                  <div>
                    <h4 className="font-bold text-neutral-900">{tx.title}</h4>
                    <p className="text-xs text-neutral-400 font-medium mt-0.5">{new Date(tx.date).toLocaleDateString()} at {new Date(tx.date).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</p>
                  </div>
               </div>
               
               <div className={`font-black text-lg ${tx.type === "credit" ? "text-green-600" : "text-neutral-900"}`}>
                 {tx.type === "credit" ? "+" : "-"}${tx.amount.toFixed(2)}
               </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
