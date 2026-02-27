export default function IssueDetailPage() {
  return (
    <section className="space-y-3 rounded-md border bg-white p-4">
      <h1 className="text-lg font-semibold">Issue Detail</h1>
      <p className="text-sm text-slate-600">Full order context + message thread</p>
      <textarea className="h-24 w-full rounded border p-2" placeholder="Thread reply" />
      <div className="grid gap-2 md:grid-cols-2">
        <input className="rounded border px-3 py-2" placeholder="Refund amount" />
        <select className="rounded border px-3 py-2">
          <option>Accountable party</option>
          <option>tenant</option>
          <option>rider</option>
          <option>platform</option>
        </select>
      </div>
      <div className="flex gap-2">
        <button className="rounded bg-emerald-600 px-3 py-2 text-sm text-white">Approve</button>
        <button className="rounded bg-rose-600 px-3 py-2 text-sm text-white">Reject</button>
      </div>
    </section>
  );
}
