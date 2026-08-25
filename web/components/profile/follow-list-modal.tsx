"use client";

import React from "react";
import Link from "next/link";
import { FollowerInfo } from "@/lib/types/profile.types";
import { Modal } from "@/components/ui/modal";

interface FollowListModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  users: FollowerInfo[];
}

export default function FollowListModal({ isOpen, onClose, title, users }: FollowListModalProps) {
  return (
    <Modal
      open={isOpen}
      onClose={onClose}
      rubric={users.length === 1 ? "1 person" : `${users.length} people`}
      title={title}
      footer={
        <button onClick={onClose} className="btn btn-quiet">
          Close
        </button>
      }
    >
      <div className="mt-4 max-h-72 divide-y divide-rule overflow-y-auto border-t border-rule">
        {users.length === 0 ? (
          <p className="meta py-8 text-center">Nobody here yet</p>
        ) : (
          users.map((u) => (
            <div key={u.username} className="flex items-center justify-between gap-3 py-3">
              <div className="min-w-0">
                <Link
                  href={`/${u.username}`}
                  onClick={onClose}
                  className="link block truncate text-sm font-medium"
                >
                  {u.name}
                </Link>
                <p className="meta truncate normal-case">@{u.username}</p>
              </div>
              {u.is_mutual && <span className="meta shrink-0">Mutual</span>}
            </div>
          ))
        )}
      </div>
    </Modal>
  );
}
