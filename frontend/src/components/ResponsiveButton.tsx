import React from "react";

interface ResponsiveButtonProps {
  onClick: () => void;
  text?: string;
  icon?: React.ReactNode;
  className?: string;
}

const ResponsiveButton: React.FC<ResponsiveButtonProps> = ({ onClick, text, icon, className }) => (
  <button className={`btn btn-primary ${className ?? ""}`} onClick={onClick}>
    {icon}
    {text && <span className="btn-responsive-text">{text}</span>}
  </button>
);

export default ResponsiveButton;
