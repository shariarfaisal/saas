"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Plus, GripVertical, Upload, Edit, Trash2 } from "lucide-react";

type Category = {
  id: string;
  name: string;
  sortOrder: number;
  isAvailable: boolean;
  productCount: number;
};

type Product = {
  id: string;
  name: string;
  price: number;
  categoryId: string;
  isAvailable: boolean;
  hasVariants: boolean;
  hasAddons: boolean;
  image?: string;
  foodType?: "veg" | "egg" | "non_veg";
};

const mockCategories: Category[] = [
  { id: "cat-1", name: "Biryani", sortOrder: 0, isAvailable: true, productCount: 8 },
  { id: "cat-2", name: "Kebabs", sortOrder: 1, isAvailable: true, productCount: 5 },
  { id: "cat-3", name: "Rice & Curry", sortOrder: 2, isAvailable: true, productCount: 12 },
  { id: "cat-4", name: "Beverages", sortOrder: 3, isAvailable: false, productCount: 6 },
  { id: "cat-5", name: "Desserts", sortOrder: 4, isAvailable: true, productCount: 4 },
];

const mockProducts: Product[] = [
  { id: "p-1", name: "Kacchi Biryani (Half)", price: 350, categoryId: "cat-1", isAvailable: true, hasVariants: true, hasAddons: true, foodType: "non_veg" },
  { id: "p-2", name: "Kacchi Biryani (Full)", price: 650, categoryId: "cat-1", isAvailable: true, hasVariants: true, hasAddons: true, foodType: "non_veg" },
  { id: "p-3", name: "Tehari", price: 220, categoryId: "cat-1", isAvailable: true, hasVariants: false, hasAddons: true, foodType: "non_veg" },
  { id: "p-4", name: "Plain Biryani", price: 180, categoryId: "cat-1", isAvailable: false, hasVariants: false, hasAddons: false, foodType: "veg" },
  { id: "p-5", name: "Chicken Reshmi Kebab", price: 280, categoryId: "cat-2", isAvailable: true, hasVariants: false, hasAddons: false, foodType: "non_veg" },
  { id: "p-6", name: "Shami Kebab", price: 120, categoryId: "cat-2", isAvailable: true, hasVariants: false, hasAddons: false, foodType: "non_veg" },
  { id: "p-7", name: "Borhani", price: 60, categoryId: "cat-4", isAvailable: true, hasVariants: true, hasAddons: false, foodType: "veg" },
];

// Emoji prefix provides visual cue; text label ensures accessibility for screen readers.
const FOOD_TYPE_LABELS: Record<string, string> = { veg: "🟢 Veg", egg: "🟡 Egg", non_veg: "🔴 Non-Veg" };
const FOOD_TYPE_VARIANTS: Record<string, "success" | "warning" | "danger"> = {
  veg: "success",
  egg: "warning",
  non_veg: "danger",
};

type NewProduct = {
  name: string;
  description: string;
  price: string;
  categoryId: string;
  foodType: "veg" | "egg" | "non_veg";
  available: boolean;
};

export default function MenuPage() {
  const [categories, setCategories] = useState(mockCategories);
  const [selectedCategoryId, setSelectedCategoryId] = useState<string>(mockCategories[0]?.id ?? "");
  const [showProductForm, setShowProductForm] = useState(false);
  const [showCsvModal, setShowCsvModal] = useState(false);
  const [showCategoryForm, setShowCategoryForm] = useState(false);
  const [newCategoryName, setNewCategoryName] = useState("");
  const [draggedCategoryId, setDraggedCategoryId] = useState<string | null>(null);
  const [newProduct, setNewProduct] = useState<NewProduct>({
    name: "",
    description: "",
    price: "",
    categoryId: mockCategories[0]?.id ?? "",
    foodType: "non_veg",
    available: true,
  });

  const filteredProducts = mockProducts.filter((p) => p.categoryId === selectedCategoryId);

  const toggleCategoryAvailability = (id: string) => {
    setCategories((prev) => prev.map((c) => (c.id === id ? { ...c, isAvailable: !c.isAvailable } : c)));
  };

  const handleDragStart = (id: string) => {
    setDraggedCategoryId(id);
  };

  const handleDragOver = (e: React.DragEvent, targetId: string) => {
    e.preventDefault();
    if (!draggedCategoryId || draggedCategoryId === targetId) return;

    setCategories((prev) => {
      const items = [...prev];
      const dragIdx = items.findIndex((c) => c.id === draggedCategoryId);
      const targetIdx = items.findIndex((c) => c.id === targetId);
      const [removed] = items.splice(dragIdx, 1);
      items.splice(targetIdx, 0, removed);
      return items.map((c, i) => ({ ...c, sortOrder: i }));
    });
  };

  const handleDragEnd = () => {
    setDraggedCategoryId(null);
    // In production, call API to persist sort order
  };

  const addCategory = () => {
    if (!newCategoryName.trim()) return;
    const newCat: Category = {
      id: `cat-new-${Date.now()}`,
      name: newCategoryName,
      sortOrder: categories.length,
      isAvailable: true,
      productCount: 0,
    };
    setCategories((prev) => [...prev, newCat]);
    setNewCategoryName("");
    setShowCategoryForm(false);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Menu Management</h1>
        <div className="flex gap-2">
          <Button className="bg-slate-600 hover:bg-slate-500" onClick={() => setShowCsvModal(true)}>
            <Upload className="mr-1 h-4 w-4" />
            Bulk Upload
          </Button>
          <Button onClick={() => setShowProductForm(true)}>
            <Plus className="mr-1 h-4 w-4" />
            Add Product
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-[260px_1fr]">
        {/* Category List */}
        <div className="rounded-md border bg-white p-3">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-sm font-semibold">Categories</h2>
            <button
              className="text-xs text-slate-600 underline"
              onClick={() => setShowCategoryForm(!showCategoryForm)}
            >
              + Add
            </button>
          </div>

          {showCategoryForm && (
            <div className="mb-3 flex gap-2">
              <Input
                value={newCategoryName}
                onChange={(e) => setNewCategoryName(e.target.value)}
                placeholder="Category name"
                className="text-sm"
              />
              <Button onClick={addCategory} className="text-xs">
                Save
              </Button>
            </div>
          )}

          <div className="space-y-1">
            {categories.map((cat) => (
              <div
                key={cat.id}
                draggable
                onDragStart={() => handleDragStart(cat.id)}
                onDragOver={(e) => handleDragOver(e, cat.id)}
                onDragEnd={handleDragEnd}
                className={`flex cursor-pointer items-center gap-2 rounded px-2 py-2 text-sm ${
                  selectedCategoryId === cat.id ? "bg-slate-900 text-white" : "hover:bg-slate-100"
                } ${draggedCategoryId === cat.id ? "opacity-50" : ""}`}
                onClick={() => setSelectedCategoryId(cat.id)}
              >
                <GripVertical className="h-3 w-3 flex-shrink-0 cursor-grab text-slate-400" />
                <span className="flex-1">{cat.name}</span>
                <span className="text-xs opacity-70">{cat.productCount}</span>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    toggleCategoryAvailability(cat.id);
                  }}
                  className={`h-3 w-3 rounded-full ${cat.isAvailable ? "bg-emerald-400" : "bg-rose-400"}`}
                  title={cat.isAvailable ? "Available" : "Unavailable"}
                />
              </div>
            ))}
          </div>
        </div>

        {/* Product Grid */}
        <div className="rounded-md border bg-white p-4">
          <h2 className="mb-3 text-sm font-semibold">
            Products — {categories.find((c) => c.id === selectedCategoryId)?.name}
          </h2>

          {filteredProducts.length === 0 ? (
            <p className="py-8 text-center text-sm text-slate-500">No products in this category.</p>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {filteredProducts.map((product) => (
                <div key={product.id} className="rounded-md border p-3">
                  <div className="flex items-start justify-between">
                    <div>
                      <p className="text-sm font-medium">{product.name}</p>
                      <p className="mt-1 text-lg font-semibold">৳{product.price}</p>
                    </div>
                    <div className="flex items-center gap-1">
                      <button className="rounded p-1 hover:bg-slate-100" title="Edit">
                        <Edit className="h-3.5 w-3.5 text-slate-500" />
                      </button>
                      <button className="rounded p-1 hover:bg-slate-100" title="Delete">
                        <Trash2 className="h-3.5 w-3.5 text-rose-500" />
                      </button>
                    </div>
                  </div>
                  <div className="mt-2 flex flex-wrap items-center gap-2">
                    {product.foodType && (
                      <Badge variant={FOOD_TYPE_VARIANTS[product.foodType]}>
                        {FOOD_TYPE_LABELS[product.foodType]}
                      </Badge>
                    )}
                    {product.hasVariants && <Badge variant="info">Variants</Badge>}
                    {product.hasAddons && <Badge variant="info">Addons</Badge>}
                    <Badge variant={product.isAvailable ? "success" : "danger"}>
                      {product.isAvailable ? "Available" : "Unavailable"}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Product Create/Edit Sheet */}
      {showProductForm && (
        <div className="fixed inset-0 z-50 flex justify-end bg-black/30" onClick={() => setShowProductForm(false)}>
          <div className="w-full max-w-lg overflow-y-auto bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="mb-4 text-lg font-semibold">Add Product</h2>
            <form
              className="space-y-4"
              onSubmit={(e) => {
                e.preventDefault();
                // In production, POST to /partner/restaurants/{id}/products
                setShowProductForm(false);
              }}
            >
              <div>
                <label className="mb-1 block text-sm font-medium">Product Name *</label>
                <Input
                  placeholder="e.g. Kacchi Biryani"
                  value={newProduct.name}
                  onChange={(e) => setNewProduct((p) => ({ ...p, name: e.target.value }))}
                  required
                />
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium">Description</label>
                <Textarea
                  placeholder="Describe this item..."
                  rows={2}
                  value={newProduct.description}
                  onChange={(e) => setNewProduct((p) => ({ ...p, description: e.target.value }))}
                />
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium">Price (৳) *</label>
                <Input
                  type="number"
                  placeholder="350"
                  value={newProduct.price}
                  onChange={(e) => setNewProduct((p) => ({ ...p, price: e.target.value }))}
                  required
                />
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium">Category</label>
                <select
                  className="w-full rounded-md border px-3 py-2 text-sm"
                  value={newProduct.categoryId}
                  onChange={(e) => setNewProduct((p) => ({ ...p, categoryId: e.target.value }))}
                >
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium">Food Type</label>
                <div className="flex gap-3">
                  {(["veg", "egg", "non_veg"] as const).map((ft) => (
                    <label key={ft} className="flex cursor-pointer items-center gap-1 text-sm">
                      <input
                        type="radio"
                        name="foodType"
                        value={ft}
                        checked={newProduct.foodType === ft}
                        onChange={() => setNewProduct((p) => ({ ...p, foodType: ft }))}
                      />
                      <Badge variant={FOOD_TYPE_VARIANTS[ft]}>{FOOD_TYPE_LABELS[ft]}</Badge>
                    </label>
                  ))}
                </div>
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium">Images</label>
                <Input type="file" accept="image/*" multiple />
              </div>

              {/* Variant Builder */}
              <div className="rounded border p-3">
                <h3 className="mb-2 text-sm font-semibold">Variants</h3>
                <p className="text-xs text-slate-500">Add size/portion variants with different prices</p>
                <div className="mt-2 space-y-2">
                  <div className="flex gap-2">
                    <Input placeholder="Variant name (e.g. Half)" className="flex-1" />
                    <Input placeholder="Price" type="number" className="w-24" />
                    <Button type="button" className="text-xs">Add</Button>
                  </div>
                </div>
              </div>

              {/* Addon Builder */}
              <div className="rounded border p-3">
                <h3 className="mb-2 text-sm font-semibold">Addons</h3>
                <p className="text-xs text-slate-500">Add optional extras customers can select</p>
                <div className="mt-2 space-y-2">
                  <div className="flex gap-2">
                    <Input placeholder="Addon name (e.g. Extra Raita)" className="flex-1" />
                    <Input placeholder="Price" type="number" className="w-24" />
                    <Button type="button" className="text-xs">Add</Button>
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="available"
                  checked={newProduct.available}
                  onChange={(e) => setNewProduct((p) => ({ ...p, available: e.target.checked }))}
                />
                <label htmlFor="available" className="text-sm font-medium">
                  Available
                </label>
              </div>

              <div className="flex gap-2">
                <Button type="submit">Save Product</Button>
                <Button
                  type="button"
                  className="bg-slate-200 text-slate-700 hover:bg-slate-300"
                  onClick={() => setShowProductForm(false)}
                >
                  Cancel
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* CSV Upload Modal */}
      {showCsvModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={() => setShowCsvModal(false)}>
          <div className="w-full max-w-md rounded-md bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="mb-4 text-lg font-semibold">Bulk Upload Products</h2>
            <p className="mb-3 text-sm text-slate-600">
              Upload a CSV file with columns: category_name, name, description, base_price, availability
            </p>
            <Input type="file" accept=".csv" />
            <div className="mt-4 flex gap-2">
              <Button>Upload & Import</Button>
              <Button
                className="bg-slate-200 text-slate-700 hover:bg-slate-300"
                onClick={() => setShowCsvModal(false)}
              >
                Cancel
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
