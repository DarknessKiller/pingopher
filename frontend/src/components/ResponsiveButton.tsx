// ResponsiveAddHostButton.tsx
import React, { useState, useEffect } from "react";
import { Button } from "antd";

interface ResponsiveButtonProps {
  onClick: () => void;
  text?: string;
  icon?: React.ReactNode;
  sizeMobile?: "small" | "middle" | "large";
  sizeDesktop?: "small" | "middle" | "large";
  className?: string;
}

const ResponsiveButton: React.FC<ResponsiveButtonProps> = ({
  onClick,
  text,
  icon,
  sizeMobile = "small",
  sizeDesktop = "middle",
  className,
}) => {
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth < 768);
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  return (
    <Button
      type="primary"
      icon={icon}
      onClick={onClick}
      size={isMobile ? sizeMobile : sizeDesktop}
      className={className}
    >
      {!isMobile && text}
    </Button>
  );
};

export default ResponsiveButton;
