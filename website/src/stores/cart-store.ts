import { create } from "zustand";
import { persist } from "zustand/middleware";

export type CartItem = {
  id: string; // Unique ID for this item in cart (includes modifiers)
  productId: string;
  name: string;
  price: number;
  quantity: number;
  image?: string;
  modifiers?: {
    groupId: string;
    groupName: string;
    optionId: string;
    optionName: string;
    price: number;
  }[];
  restaurantId: string;
  restaurantName: string;
};

type CartState = {
  items: CartItem[];
  addItem: (item: CartItem) => void;
  removeItem: (itemId: string) => void;
  updateQuantity: (itemId: string, quantity: number) => void;
  clearCart: () => void;
  totalItems: () => number;
  subtotal: () => number;
};

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      items: [],
      
      addItem: (newItem) => {
        const items = get().items;
        const existingItemIndex = items.findIndex((i) => i.id === newItem.id);

        if (existingItemIndex > -1) {
          const updatedItems = [...items];
          updatedItems[existingItemIndex].quantity += newItem.quantity;
          set({ items: updatedItems });
        } else {
          set({ items: [...items, newItem] });
        }
      },

      removeItem: (itemId) => {
        set({ items: get().items.filter((i) => i.id !== itemId) });
      },

      updateQuantity: (itemId, quantity) => {
        if (quantity < 1) {
          get().removeItem(itemId);
          return;
        }
        const updatedItems = get().items.map((i) =>
          i.id === itemId ? { ...i, quantity } : i
        );
        set({ items: updatedItems });
      },

      clearCart: () => set({ items: [] }),

      totalItems: () => get().items.reduce((sum, item) => sum + item.quantity, 0),

      subtotal: () =>
        get().items.reduce(
          (sum, item) => sum + (item.price + (item.modifiers?.reduce((s, m) => s + m.price, 0) || 0)) * item.quantity,
          0
        ),
    }),
    {
      name: "munchies-cart",
    }
  )
);
