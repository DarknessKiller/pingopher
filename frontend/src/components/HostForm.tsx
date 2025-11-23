import React, { useEffect } from "react";
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

const { Option } = Select;

interface HostFormProps {
  initialValues?: Host;
  onSuccess: () => void;
}

const HostForm: React.FC<HostFormProps> = ({ initialValues, onSuccess }) => {
  const [form] = Form.useForm<CreateHostRequest>();

  useEffect(() => {
    if (initialValues) {
      form.setFieldsValue(initialValues);
    } else {
      form.resetFields();
    }
  }, [initialValues, form]);

  const [loading, setLoading] = React.useState(false);

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
    } catch (error) {
      message.error((error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={onFinish}
      initialValues={{
        protocol: "https",
        pingInterval: 60,
        failThreshold: 3,
      }}
    >
      <Row gutter={16}>
        <Col span={18}>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: "Please enter host name" }]}
          >
            <Input placeholder="My Server" />
          </Form.Item>
        </Col>

        <Col span="6">
          <Form.Item
            name="protocol"
            label="Protocol"
            rules={[{ required: true, message: "Please select protocol" }]}
          >
            <Select>
              <Option value="http">HTTP</Option>
              <Option value="https">HTTPS</Option>
            </Select>
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={18}>
          <Form.Item
            name="hostUrl"
            label="Host URL/IP"
            rules={[{ required: true, message: "Please enter host URL or IP" }]}
          >
            <Input placeholder="example.com" />
          </Form.Item>
        </Col>

        <Col span={6}>
          <Form.Item name="port" label="Port">
            <InputNumber min={0} max={65535} style={{ width: "100%" }} />
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            name="pingInterval"
            label="Ping Interval (seconds)"
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
                  alignItems: "flex-start",
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
                    <Select placeholder="Protocol">
                      <Option value="udp">UDP</Option>
                      <Option value="tcp">TCP</Option>
                    </Select>
                  </Form.Item>
                  <Form.Item
                    {...restField}
                    name={[name, "port"]}
                    rules={[{ required: true, message: "Missing Port" }]}
                    label="Port"
                  >
                    <InputNumber style={{ width: "100%" }} placeholder="53" />
                  </Form.Item>
                </div>
                <Button
                  type="text"
                  danger
                  onClick={() => remove(name)}
                  icon={<span style={{ fontSize: 20 }}>-</span>}
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
