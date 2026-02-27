import Link from "next/link";

export default function IssuesPage() {
  return (
    <section className="space-y-3 rounded-md border bg-white p-4">
      <h1 className="text-lg font-semibold">Issues Queue</h1>
      <div className="grid gap-2 md:grid-cols-4">
        <input className="rounded border px-3 py-2" placeholder="Status" />
        <input className="rounded border px-3 py-2" placeholder="Tenant" />
        <input className="rounded border px-3 py-2" placeholder="Type" />
      </div>
      <table className="w-full text-sm">
        <thead>
          <tr>
            <th className="text-left">Issue</th>
            <th className="text-left">Tenant</th>
            <th className="text-left">Status</th>
          </tr>
        </thead>
        <tbody>
          <tr className="border-t">
            <td className="py-2">
              <Link className="underline" href="/issues/issue-1">
                ISS-5001
              </Link>
            </td>
            <td>Kacchi Bhai</td>
            <td>open</td>
          </tr>
        </tbody>
      </table>
      <div className="rounded border p-3 text-sm">Resolved issues history: 321 cases (last 30 days)</div>
    </section>
  );
}
