"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Card, CardTitle } from "@/components/ui/card";
import { Plus, Mail, Trash2 } from "lucide-react";

type TeamMember = {
  id: string;
  name: string;
  email: string;
  role: "owner" | "admin" | "manager" | "staff";
  status: "active" | "pending";
  joinedAt: string;
};

const mockTeam: TeamMember[] = [
  { id: "tm-1", name: "Faisal Ahmed", email: "faisal@kacchibhai.com", role: "owner", status: "active", joinedAt: "2024-01-15" },
  { id: "tm-2", name: "Rashid Khan", email: "rashid@kacchibhai.com", role: "admin", status: "active", joinedAt: "2024-03-20" },
  { id: "tm-3", name: "Samira Haque", email: "samira@kacchibhai.com", role: "manager", status: "active", joinedAt: "2024-06-10" },
  { id: "tm-4", name: "Pending Invite", email: "newmember@email.com", role: "staff", status: "pending", joinedAt: "2024-12-28" },
];

const roleLabels = { owner: "Owner", admin: "Admin", manager: "Manager", staff: "Staff" };
const roleVariants = { owner: "danger" as const, admin: "info" as const, manager: "warning" as const, staff: "default" as const };

export default function TeamPage() {
  const [team, setTeam] = useState(mockTeam);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("staff");

  const handleInvite = () => {
    if (!inviteEmail) return;
    const newMember: TeamMember = {
      id: `tm-new-${Date.now()}`,
      name: "Pending Invite",
      email: inviteEmail,
      role: inviteRole as TeamMember["role"],
      status: "pending",
      joinedAt: new Date().toISOString().split("T")[0],
    };
    setTeam((prev) => [...prev, newMember]);
    setInviteEmail("");
    setShowInviteModal(false);
  };

  const removeMember = (id: string) => {
    setTeam((prev) => prev.filter((m) => m.id !== id));
  };

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold">Team Management</h1>
        <Button onClick={() => setShowInviteModal(true)}>
          <Plus className="mr-1 h-4 w-4" />
          Invite Member
        </Button>
      </div>

      <Card>
        <CardTitle>Team Members</CardTitle>
        <div className="mt-4 space-y-3">
          {team.map((member) => (
            <div key={member.id} className="flex items-center gap-4 rounded-md border p-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-slate-200 font-semibold text-slate-600">
                {member.name.charAt(0).toUpperCase()}
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <p className="font-medium">{member.name}</p>
                  {member.status === "pending" && <Badge variant="warning">Pending</Badge>}
                </div>
                <p className="text-xs text-slate-500">{member.email}</p>
              </div>
              <Badge variant={roleVariants[member.role]}>{roleLabels[member.role]}</Badge>
              {member.role !== "owner" && (
                <button
                  onClick={() => removeMember(member.id)}
                  className="rounded p-1 hover:bg-slate-100"
                  title="Remove member"
                >
                  <Trash2 className="h-4 w-4 text-rose-500" />
                </button>
              )}
            </div>
          ))}
        </div>
      </Card>

      {/* Role Descriptions */}
      <Card>
        <CardTitle>Role Permissions</CardTitle>
        <div className="mt-4 space-y-3 text-sm">
          <div className="border-b pb-2">
            <p className="font-medium">Owner</p>
            <p className="text-xs text-slate-500">Full access to all features, billing, and team management</p>
          </div>
          <div className="border-b pb-2">
            <p className="font-medium">Admin</p>
            <p className="text-xs text-slate-500">All features except billing and owner transfer</p>
          </div>
          <div className="border-b pb-2">
            <p className="font-medium">Manager</p>
            <p className="text-xs text-slate-500">Order management, menu editing, reports viewing, rider management</p>
          </div>
          <div>
            <p className="font-medium">Staff</p>
            <p className="text-xs text-slate-500">Order acceptance/rejection, basic order viewing only</p>
          </div>
        </div>
      </Card>

      {/* Invite Modal */}
      {showInviteModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={() => setShowInviteModal(false)}>
          <div className="w-full max-w-md rounded-md bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="mb-4 text-lg font-semibold">Invite Team Member</h2>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium">Email Address</label>
                <div className="flex items-center gap-2">
                  <Mail className="h-4 w-4 text-slate-400" />
                  <Input
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    placeholder="member@email.com"
                    type="email"
                  />
                </div>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Role</label>
                <Select value={inviteRole} onChange={(e) => setInviteRole(e.target.value)}>
                  <option value="admin">Admin</option>
                  <option value="manager">Manager</option>
                  <option value="staff">Staff</option>
                </Select>
              </div>
              <div className="flex gap-2">
                <Button onClick={handleInvite}>Send Invitation</Button>
                <Button
                  className="bg-slate-200 text-slate-700 hover:bg-slate-300"
                  onClick={() => setShowInviteModal(false)}
                >
                  Cancel
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
