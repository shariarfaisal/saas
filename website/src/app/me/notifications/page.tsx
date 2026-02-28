"use client";

import React, { useState } from "react";

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState([
    { id: "1", type: "promo", title: "50% OFF your next order!", message: "Use code HUNGRY50 at checkout. Valid until midnight.", date: "2 Hours ago", read: false },
    { id: "2", type: "order", title: "Order Delivered", message: "Your order #M1029 has been delivered. Enjoy your meal!", date: "Yesterday", read: true },
    { id: "3", type: "system", title: "Welcome to Munchies", message: "Thanks for joining us. Check out the best restaurants around you.", date: "Oct 10", read: true },
  ]);

  const markAllAsRead = () => {
    setNotifications(notifications.map(n => ({ ...n, read: true })));
  };

  return (
    <div className="bg-white rounded-3xl p-6 sm:p-8 shadow-sm border border-neutral-100">
      <div className="flex items-center justify-between mb-8">
        <h2 className="text-2xl font-bold text-neutral-900">Notifications</h2>
        <button 
          onClick={markAllAsRead}
          className="text-sm font-bold text-orange-600 hover:text-orange-700 transition-colors"
        >
          Mark all as read
        </button>
      </div>

      {notifications.length === 0 ? (
         <div className="text-center py-16 border-2 border-dashed border-neutral-100 rounded-3xl">
            <div className="w-16 h-16 bg-neutral-50 rounded-full flex mx-auto items-center justify-center text-3xl mb-4">🔔</div>
            <p className="text-neutral-500 font-medium mb-6">You&apos;re all caught up!</p>
         </div>
      ) : (
        <div className="space-y-3">
          {notifications.map(notification => (
            <div key={notification.id} className={`p-4 rounded-2xl flex gap-4 transition-all ${notification.read ? 'bg-neutral-50 border border-neutral-100/50 opacity-70' : 'bg-white border border-neutral-200 shadow-sm'}`}>
              <div className={`w-12 h-12 rounded-full flex items-center justify-center text-xl shrink-0 ${notification.type === 'promo' ? 'bg-purple-100 text-purple-600' : notification.type === 'order' ? 'bg-green-100 text-green-600' : 'bg-blue-100 text-blue-600'}`}>
                {notification.type === 'promo' ? '🎫' : notification.type === 'order' ? '🍔' : '👋'}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex justify-between items-start mb-1 gap-2">
                   <h3 className={`font-bold truncate ${notification.read ? 'text-neutral-700' : 'text-neutral-900'}`}>{notification.title}</h3>
                   <span className="text-[10px] font-bold text-neutral-400 whitespace-nowrap pt-1 uppercase tracking-wider">{notification.date}</span>
                </div>
                <p className="text-sm text-neutral-500 leading-relaxed font-medium">
                  {notification.message}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
