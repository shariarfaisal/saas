export async function fetchClientApi<T>(path: string, options: RequestInit = {}): Promise<T> {
  const base = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
  const authHeaders: Record<string, string> = {};
  if (typeof window !== "undefined") {
    const token = localStorage.getItem('token');
    if (token) authHeaders["Authorization"] = `Bearer ${token}`;
  }
  const res = await fetch(`${base}${path}`, {
    ...options,
    headers: {
      ...options.headers,
      "Content-Type": "application/json",
      ...authHeaders,
    }
  });
  if (!res.ok) {
    let errMessage = "Failed to fetch from API";
    try {
      const errBody = await res.json();
      errMessage = errBody?.error?.message || errBody?.message || errMessage;
    } catch {
      // Ignore
    }
    throw new Error(`API Error (${res.status}): ${errMessage}`);
  }
  return res.json() as Promise<T>;
}

// --- Types ---
export type Banner = {
  id: string;
  title: string;
  subtitle?: string;
  image_url: string;
  mobile_image_url?: string;
  link_type?: string;
  link_value?: string;
};

export type Story = {
  id: string;
  title?: string;
  media_url: string;
  media_type: string;
  thumbnail_url?: string;
  link_type?: string;
  link_value?: string;
};

export type Section = {
  id: string;
  title: string;
  subtitle?: string;
  content_type: string;
  // Based on your specific implementation, it might contain nested items. We'll simplify.
  items?: Record<string, unknown>[];
};

export type Area = {
  id: string;
  name: string;
  slug: string;
};

export type Restaurant = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  cover_image?: string;
  logo_url?: string;
  rating?: number;
  delivery_time_mins?: number;
  is_open?: boolean;
  cuisines?: string[];
  has_discount?: boolean;
  discount_price?: string;
};

export type PagedResponse<T> = {
  data: T[];
  meta: {
    total: number;
    page: number;
    per_page: number;
  };
};

export type Product = {
  id: string;
  name: string;
  description?: string;
  price: number; // Changed to number to match backend NUMERIC which json unmarshals to number
  image_url?: string;
  is_available: boolean;
  category_id?: string;
  has_modifiers: boolean;
  has_discount?: boolean;
  discount_price?: number;
};

export type Category = {
  id: string;
  name: string;
  sort_order: number;
};

export type ModifierOption = {
  id: string;
  name: string;
  additional_price: number;
  is_available: boolean;
};

export type ModifierGroup = {
  id: string;
  name: string;
  min_required: number;
  max_allowed: number;
  options: ModifierOption[];
};

export type Discount = {
  id: string;
  discount_type: 'percentage' | 'fixed';
  amount: number;
  max_discount_cap?: number;
};

export type ProductDetail = Product & {
  modifier_groups?: ModifierGroup[];
  discount?: Discount;
};
export type OrderItem = {
  id: string;
  product_id: string;
  product_name: string;
  quantity: number;
  unit_price: number;
  modifier_price: number;
  item_total: number;
  selected_modifiers?: Record<string, unknown>[];
};

export type OrderTimelineEvent = {
  id: string;
  event_type: string;
  new_status?: string;
  description: string;
  created_at: string;
};

export type Order = {
  id: string;
  order_number: string;
  status: string;
  payment_status: string;
  payment_method: string;
  total_amount: number;
  subtotal: number;
  delivery_charge: number;
  item_discount_total: number;
  promo_discount_total: number;
  delivery_address: Record<string, unknown>;
  delivery_recipient_name: string;
  delivery_recipient_phone: string;
  delivery_area: string;
  created_at: string;
};

export type OrderDetail = {
  order: Order;
  items: OrderItem[];
  timeline: OrderTimelineEvent[];
};

export type Address = {
  id: string;
  name: string;
  address_line?: string;
  area: string;
  is_default: boolean;
  latitude?: string;
  longitude?: string;
};

export type CalculateChargesRequest = {
  items: {
    product_id: string;
    restaurant_id: string;
    category_id?: string;
    quantity: number;
    unit_price: string;
    modifier_price: string;
    item_discount: string;
    item_vat: string;
    product_name: string;
  }[];
  promo_code?: string;
  delivery_area: string;
};

export type ChargeBreakdown = {
  subtotal: number;
  item_discount_total: number;
  promo_discount_total: number;
  vat_total: number;
  delivery_charge: number;
  service_fee: number;
  total_amount: number;
  promo_result?: {
    valid: boolean;
    discount_amount: number;
    error_message?: string;
    code?: string;
  };
};

export type CreateOrderRequest = {
  items: {
    product_id: string;
    restaurant_id: string;
    category_id?: string;
    quantity: number;
    unit_price: string;
    modifier_price: string;
    product_name: string;
    product_snapshot: Record<string, unknown>;
    selected_modifiers: Record<string, unknown>[];
    item_discount: string;
    item_vat: string;
  }[];
  promo_code?: string;
  payment_method: string;
  delivery_address_id?: string;
  delivery_recipient_name: string;
  delivery_recipient_phone: string;
  delivery_area: string;
  delivery_geo_lat?: string;
  customer_note?: string;
};

export type CreateOrderResponse = {
  payment_url?: string;
  order?: {
    id?: string;
    ID?: string;
    order_number?: string;
    [key: string]: unknown;
  };
  [key: string]: unknown;
};

/**
 * Calculate order charges
 */
export async function calculateCharges(req: CalculateChargesRequest): Promise<ChargeBreakdown> {
  const isServer = typeof window === "undefined";
  if (isServer) {
    return fetchClientApi<ChargeBreakdown>("/orders/charges/calculate", {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}/orders/charges/calculate`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // Host will be picked up by proxy or we rely on relative paths if frontend handles it
    },
    body: JSON.stringify(req),
  });
  if (!res.ok) throw new Error("Failed to calculate charges");
  return res.json();
}

/**
 * Create order
 */
export async function createOrder(req: CreateOrderRequest): Promise<CreateOrderResponse> {
    const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}/orders`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${localStorage.getItem('token')}` // Placeholder for token
    },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
      const err = await res.json();
      throw new Error(err.message || "Failed to create order");
  }
  return res.json();
}

/**
 * Fetch my orders
 */
export async function fetchMyOrders(page = 1, perPage = 10): Promise<PagedResponse<Order>> {
  const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}/me/orders?page=${page}&per_page=${perPage}`, {
    headers: {
      "Authorization": `Bearer ${localStorage.getItem('token')}`
    },
  });
  if (!res.ok) throw new Error("Failed to fetch orders");
  return res.json();
}
