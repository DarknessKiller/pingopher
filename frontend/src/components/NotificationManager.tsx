import React, { useEffect, useState } from "react";
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  message,
  Space,
  Tag,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { PlusOutlined, EditOutlined, DeleteOutlined } from "@ant-design/icons";
import {
  getNotifications,
  createNotification,
  updateNotification,
  deleteNotification,
  type Host,
  type Notification,
  type CreateNotificationRequest,
} from "../api";

const { Option } = Select;

interface NotificationManagerProps {
  host: Host;
}

const NotificationManager: React.FC<NotificationManagerProps> = ({ host }) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(false);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingNotification, setEditingNotification] = useState<
    Notification | undefined
  >(undefined);
  const [form] = Form.useForm<CreateNotificationRequest>();

  const fetchNotifications = React.useCallback(async () => {
    setLoading(true);
    try {
      const response = await getNotifications(host.id);
      setNotifications(response.data || []);
    } catch (error) {
      message.error((error as Error).message);
    } finally {
      setLoading(false);
    }
  }, [host.id]);

  useEffect(() => {
    if (host) {
      fetchNotifications();
    }
  }, [host, fetchNotifications]);

  const handleCreate = () => {
    setEditingNotification(undefined);
    form.resetFields();
    form.setFieldsValue({ type: "discord", active: true });
    setIsModalVisible(true);
  };

  const handleEdit = (record: Notification) => {
    setEditingNotification(record);
    form.setFieldsValue(record);
    setIsModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteNotification(host.id, id);
      message.success("Notification deleted");
      fetchNotifications();
    } catch (error) {
      message.error((error as Error).message);
    }
  };

  const onFinish = async (values: CreateNotificationRequest) => {
    try {
      if (editingNotification) {
        await updateNotification(host.id, editingNotification.id, values);
        message.success("Notification updated");
      } else {
        await createNotification(host.id, values);
        message.success("Notification created");
      }
      setIsModalVisible(false);
      fetchNotifications();
    } catch (error) {
      message.error((error as Error).message);
    }
  };

  const columns: ColumnsType<Notification> = [
    {
      title: "Name",
      dataIndex: "name",
      key: "name",
    },
    {
      title: "Type",
      dataIndex: "type",
      key: "type",
      render: (type: string) => <Tag color="blue">{type.toUpperCase()}</Tag>,
    },
    {
      title: "Active",
      dataIndex: "active",
      key: "active",
      render: (active: boolean) => (
        <Tag color={active ? "green" : "red"}>
          {active ? "Active" : "Inactive"}
        </Tag>
      ),
    },
    {
      title: "Actions",
      key: "actions",
      align: "right",
      render: (_, record) => (
        <Space>
          <Button icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          <Button
            icon={<DeleteOutlined />}
            danger
            onClick={() =>
              Modal.confirm({
                title: "Are you sure?",
                onOk: () => handleDelete(record.id),
              })
            }
          />
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div
        style={{
          marginBottom: 16,
          display: "flex",
          justifyContent: "flex-end",
          alignItems: "center",
        }}
      >
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          Add Notification
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={notifications}
        rowKey="id"
        loading={loading}
        pagination={false}
        scroll={{ x: "max-content" }}
      />

      <Modal
        title={editingNotification ? "Edit Notification" : "Add Notification"}
        open={isModalVisible}
        onCancel={() => setIsModalVisible(false)}
        onOk={() => form.submit()}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="My Discord Alert" />
          </Form.Item>
          <Form.Item name="type" label="Type" rules={[{ required: true }]}>
            <Select>
              <Option value="discord">Discord</Option>
            </Select>
          </Form.Item>
          <Form.Item name="active" label="Active" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item
            noStyle
            shouldUpdate={(prev, current) => prev.type !== current.type}
          >
            {({ getFieldValue }) =>
              getFieldValue("type") === "discord" ? (
                <>
                  <Form.Item
                    name="discordWebhookUrl"
                    label="Webhook URL"
                    rules={[{ required: true, type: "url" }]}
                  >
                    <Input />
                  </Form.Item>
                  <Form.Item name="discordUsername" label="Username">
                    <Input placeholder="Pingopher" />
                  </Form.Item>
                  <Form.Item name="discordPrefixMessage" label="Prefix Message">
                    <Input.TextArea placeholder="@here" />
                  </Form.Item>
                </>
              ) : null
            }
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default NotificationManager;
