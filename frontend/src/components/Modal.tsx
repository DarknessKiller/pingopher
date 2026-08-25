import React, { useEffect, useState } from "react";
import { CloseIcon } from "./icons";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  width?: number;
  footer?: React.ReactNode;
  zIndex?: number;
}

const Modal: React.FC<ModalProps> = ({ open, onClose, title, children, width, footer, zIndex }) => {
  const [closing, setClosing] = useState(false);

  useEffect(() => {
    if (!open) setClosing(false);
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, onClose]);

  if (!open && !closing) return null;

  const handleClose = () => {
    setClosing(true);
    setTimeout(onClose, 150);
  };

  const overlayStyle = zIndex ? { zIndex } : undefined;

  return (
    <div className={`modal-overlay${closing ? " closing" : ""}`} style={overlayStyle} onClick={handleClose}>
      <div
        className="modal-content"
        style={width ? { maxWidth: width } : undefined}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="modal-header">
          <h3>{title}</h3>
          <button className="btn btn-ghost btn-icon btn-sm" onClick={handleClose}>
            <CloseIcon size={18} />
          </button>
        </div>
        <div className="modal-body">{children}</div>
        {footer && <div className="modal-footer">{footer}</div>}
      </div>
    </div>
  );
};

export default Modal;
