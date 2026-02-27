export default function PayoutsPage() {
  return (
    <section className="rounded-md border bg-white p-4">
      <h1 className="text-lg font-semibold">Payout Tracking</h1>
      <table className="mt-2 w-full text-sm">
        <thead>
          <tr>
            <th className="text-left">Tenant</th>
            <th className="text-left">Amount</th>
            <th className="text-left">Settlement status</th>
          </tr>
        </thead>
        <tbody>
          <tr className="border-t">
            <td className="py-2">Pizza Hub</td>
            <td>৳81,250</td>
            <td>processing</td>
          </tr>
        </tbody>
      </table>
    </section>
  );
}
