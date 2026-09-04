import { useState } from "react"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"

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
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        <Button size="sm" variant="destructive">Kill</Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Kill {name}?</AlertDialogTitle>
          <AlertDialogDescription>
            {isAgent
              ? "Code OS will send SIGTERM to this agent. Its active task may stop immediately, but unrelated services are left untouched."
              : "Portly will stop this application and its managed process tree. You can start it again from Portly."}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={pending}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={pending}
            onClick={(event) => {
              event.preventDefault()
              void confirm()
            }}
          >
            {pending ? <Spinner data-icon="inline-start" /> : null}
            {pending ? "Stopping…" : "Confirm kill"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
