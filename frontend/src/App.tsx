import React, { useState, useEffect } from "react";
import { Layout, Typography, ConfigProvider, Switch, theme } from "antd";
import HostList from "./components/HostList";

const { Header, Content, Footer } = Layout;
const { Title } = Typography;

const App: React.FC = () => {
  const { darkAlgorithm, defaultAlgorithm } = theme;

  // Detect browser preference on first load
  const [darkMode, setDarkMode] = useState(() => {
    if (typeof window !== "undefined" && window.matchMedia) {
      return window.matchMedia("(prefers-color-scheme: dark)").matches;
    }
    return false;
  });

  useEffect(() => {
    if (!window.matchMedia) return;

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = (e: MediaQueryListEvent) => setDarkMode(e.matches);
    mediaQuery.addEventListener("change", handleChange);
    return () => mediaQuery.removeEventListener("change", handleChange);
  }, []);

  return (
    <ConfigProvider
      theme={{
        algorithm: darkMode ? darkAlgorithm : defaultAlgorithm,
        token: {
          colorPrimary: "#1890ff",
        },
      }}
    >
      <Layout className="layout" style={{ minHeight: "100vh" }}>
        <Header
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            flexWrap: "wrap",
            padding: "0 16px",
          }}
        >
          <Title
            level={3}
            style={{
              color: "white",
              margin: 0,
              flex: "1 1 auto",
            }}
          >
            Pingopher
          </Title>
          <Switch
            checked={darkMode}
            onChange={setDarkMode}
            checkedChildren="Dark"
            unCheckedChildren="Light"
            style={{ marginTop: 8, flex: "0 0 auto" }}
          />
        </Header>

        <Content style={{ padding: "1.5rem", marginTop: 24 }}>
          <div
            className="site-layout-content"
            style={{
              background: darkMode ? "#141414" : "#fff",
              padding: 24,
              minHeight: 380,
              width: "100%",
              boxSizing: "border-box",
            }}
          >
            <HostList />
          </div>
        </Content>

        <Footer
          style={{
            textAlign: "center",
            color: darkMode ? "white" : "black",
          }}
        >
          Pingopher ©{new Date().getFullYear()} Created by DarknessKiller
        </Footer>
      </Layout>
    </ConfigProvider>
  );
};

export default App;
