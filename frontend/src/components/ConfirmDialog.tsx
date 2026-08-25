import React, { useEffect, useState } from "react";

interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
}

const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  title,
  message,
  confirmLabel = "Delete",
  onConfirm,
  onCancel,
}) => {
  const [closing, setClosing] = useState(false);

  // Reset the closing animation state when the dialog is re-opened.
  const [prevOpen, setPrevOpen] = useState(open);
  if (open !== prevOpen) {
    setPrevOpen(open);
    setClosing(false);
  }

  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCancel();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, onCancel]);

  if (!open && !closing) return null;

  const handleClose = () => {
    setClosing(true);
    setTimeout(onCancel, 150);
  };

  return (
    <div className={`modal-overlay${closing ? " closing" : ""}`} onClick={handleClose}>
      <div
        className="confirm-dialog"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="confirm-dialog-body">
          <h3 className="confirm-dialog-title">{title}</h3>
          <p className="confirm-dialog-message">{message}</p>
        </div>
        <div className="confirm-dialog-actions">
          <button className="btn confirm-btn-cancel" onClick={handleClose}>Cancel</button>
          <button className="btn confirm-btn-danger" onClick={onConfirm}>{confirmLabel}</button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmDialog;
