import React, { useEffect, useState } from "react";
import {
  Form,
  Input,
  InputNumber,
  Select,
  Button,
  message,
  Row,
  Col,
} from "antd";
import {
  createHost,
  updateHost,
  type Host,
  type CreateHostRequest,
} from "../api";

interface HostFormProps {
  initialValues?: Host;
  onSuccess: () => void;
}

const DEFAULT_VALUES: Partial<CreateHostRequest> = {
  protocol: "https",
  pingInterval: 60,
  failThreshold: 3,
  dns: [],
};

const HostForm: React.FC<HostFormProps> = ({ initialValues, onSuccess }) => {
  const [form] = Form.useForm<CreateHostRequest>();
  const [loading, setLoading] = useState(false);

  // Sync form values when initialValues change
  useEffect(() => {
    if (initialValues) {
      form.setFieldsValue({ ...DEFAULT_VALUES, ...initialValues });
    } else {
      form.resetFields();
      form.setFieldsValue(DEFAULT_VALUES);
    }
  }, [initialValues, form]);

  // Async submit handler (React-standard)
  const onFinish = async (values: CreateHostRequest) => {
    setLoading(true);
    try {
      if (initialValues) {
        await updateHost(initialValues.id, values);
        message.success("Host updated successfully");
      } else {
        await createHost(values);
        message.success("Host created successfully");
      }
      onSuccess();
    } catch (err) {
      message.error((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Form<CreateHostRequest>
      form={form}
      layout="vertical"
      onFinish={onFinish}
      initialValues={DEFAULT_VALUES}
    >
      {/* Row 1 */}
      <Row gutter={16}>
        <Col span={16}>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: "Please enter host name" }]}
          >
            <Input placeholder="My Server" />
          </Form.Item>
        </Col>

        <Col span={8}>
          <Form.Item
            name="protocol"
            label="Protocol"
            rules={[{ required: true, message: "Please select protocol" }]}
          >
            <Select
              options={[
                { label: "HTTP", value: "http" },
                { label: "HTTPS", value: "https" },
              ]}
            />
          </Form.Item>
        </Col>
      </Row>

      {/* Row 2 */}
      <Row gutter={16}>
        <Col span={16}>
          <Form.Item
            name="hostUrl"
            label="Host URL/IP"
            rules={[{ required: true, message: "Please enter host URL or IP" }]}
          >
            <Input placeholder="example.com" />
          </Form.Item>
        </Col>

        <Col span={8}>
          <Form.Item name="port" label="Port">
            <InputNumber min={0} max={65535} style={{ width: "100%" }} />
          </Form.Item>
        </Col>
      </Row>

      {/* Row 3 */}
      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            name="pingInterval"
            label="Ping Interval"
            rules={[{ required: true, message: "Please enter ping interval" }]}
          >
            <InputNumber min={1} style={{ width: "100%" }} />
          </Form.Item>
        </Col>

        <Col span={12}>
          <Form.Item
            name="failThreshold"
            label="Fail Threshold"
            rules={[{ required: true, message: "Please enter fail threshold" }]}
          >
            <InputNumber min={1} style={{ width: "100%" }} />
          </Form.Item>
        </Col>
      </Row>

      {/* DNS List */}
      <Form.List name="dns">
        {(fields, { add, remove }) => (
          <>
            {fields.map(({ key, name, ...restField }) => (
              <div
                key={key}
                style={{
                  display: "flex",
                  gap: 8,
                  marginBottom: 8,
                  border: "1px solid #d9d9d9",
                  padding: 10,
                  borderRadius: 4,
                }}
              >
                <div style={{ flex: 1 }}>
                  <Form.Item
                    {...restField}
                    name={[name, "name"]}
                    rules={[{ required: true, message: "Missing DNS Name" }]}
                    label="DNS Name"
                  >
                    <Input placeholder="Google" />
                  </Form.Item>

                  <Form.Item
                    {...restField}
                    name={[name, "ip"]}
                    rules={[{ required: true, message: "Missing IP" }]}
                    label="IP Address"
                  >
                    <Input placeholder="8.8.8.8" />
                  </Form.Item>
                </div>

                <div style={{ flex: 1 }}>
                  <Form.Item
                    {...restField}
                    name={[name, "protocol"]}
                    rules={[{ required: true, message: "Missing Protocol" }]}
                    label="Protocol"
                  >
                    <Select
                      options={[
                        { label: "UDP", value: "udp" },
                        { label: "TCP", value: "tcp" },
                      ]}
                    />
                  </Form.Item>

                  <Form.Item
                    {...restField}
                    name={[name, "port"]}
                    rules={[{ required: true, message: "Missing Port" }]}
                    label="Port"
                  >
                    <InputNumber min={1} style={{ width: "100%" }} />
                  </Form.Item>
                </div>

                <Button
                  type="text"
                  danger
                  onClick={() => remove(name)}
                  icon={<span style={{ fontSize: 20, lineHeight: 1 }}>−</span>}
                />
              </div>
            ))}

            <Form.Item>
              <Button
                type="dashed"
                onClick={() => add()}
                block
                icon={<span>+</span>}
              >
                Add DNS Server
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>

      <Form.Item style={{ marginBottom: 0 }}>
        <Button type="primary" htmlType="submit" block loading={loading}>
          {initialValues ? "Update Host" : "Create Host"}
        </Button>
      </Form.Item>
    </Form>
  );
};

export default HostForm;
