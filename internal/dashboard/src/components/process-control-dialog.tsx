import { useState } from "react"

import { Button } from "@/components/ui/button"
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogTitle, DialogTrigger } from "@/components/ui/dialog"

type ProcessControlDialogProps = Readonly<{
  action: "application" | "agent"
  name: string
  pending: boolean
  onConfirm: () => Promise<unknown>
}>

export function ProcessControlDialog({ action, name, pending, onConfirm }: ProcessControlDialogProps) {
  const [open, setOpen] = useState(false)
  const isAgent = action === "agent"

  async function confirm() {
    await onConfirm()
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm" variant="destructive">Kill</Button>
      </DialogTrigger>
      <DialogContent className="max-w-lg">
        <div className="space-y-5 pr-10">
          <div className="space-y-2">
            <p className="font-mono text-[10px] uppercase tracking-[0.16em] text-[#e00]">Destructive action</p>
            <DialogTitle>Kill {name}?</DialogTitle>
            <DialogDescription>
              {isAgent
                ? "Code OS will send SIGTERM to this agent. Its active task may stop immediately, but unrelated services are left untouched."
                : "Portly will stop this application and its managed process tree. You can start it again from Portly."}
            </DialogDescription>
          </div>
          <div className="flex justify-end gap-2 border-t border-[#333] pt-4">
            <DialogClose asChild><Button variant="ghost" disabled={pending}>Cancel</Button></DialogClose>
            <Button variant="destructive" disabled={pending} onClick={() => void confirm()}>
              {pending ? "Stopping…" : "Confirm kill"}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
