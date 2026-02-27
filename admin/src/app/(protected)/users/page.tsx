export default function UsersPage() {
  return (
    <div className="space-y-4 rounded-md border bg-white p-4">
      <h1 className="text-lg font-semibold">User Management</h1>
      <div className="grid gap-2 md:grid-cols-4">
        <input className="rounded border px-3 py-2" placeholder="Phone/email/name" />
        <input className="rounded border px-3 py-2" placeholder="Tenant filter" />
        <button className="rounded bg-slate-900 px-3 py-2 text-sm text-white">Search</button>
      </div>
      <div className="rounded border p-3">
        <h2 className="font-medium">User Detail Drawer</h2>
        <p className="text-sm text-slate-600">Profile + order summary + status</p>
        <textarea className="mt-2 w-full rounded border p-2" placeholder="Suspend reason" />
        <div className="mt-2 flex gap-2">
          <button className="rounded bg-rose-600 px-3 py-2 text-sm text-white">Suspend User</button>
          <button className="rounded bg-slate-700 px-3 py-2 text-sm text-white">GDPR Delete (confirm)</button>
        </div>
        <p className="mt-2 text-xs text-slate-500">Data wipe status: pending verification</p>
      </div>
    </div>
  );
}
