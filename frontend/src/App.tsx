/* eslint-disable @typescript-eslint/no-explicit-any */
import "normalize.css";
import {
  List,
  Avatar,
  Empty,
  Modal,
  Input,
  notification,
  message,
  Progress,
  Flex,
} from "antd";
import {
  User,
  Resp,
  methodFreshUsers,
  SendMsg,
  Req,
  methodSendMsg,
  RegisterResp,
  methodReceiveMsg,
  ReceiveMsg,
  methodSendFileStart,
  methodReceiveFileConfirm,
  ReceiveFileConfirm,
  methodReceiveFileREFUSED,
  SendFileStartResp,
  TransferFileCancel,
  methodSendFileCancel,
} from "./model";
import {
  FileOutlined,
  MessageOutlined,
  CopyOutlined,
  DownloadOutlined,
  UploadOutlined,
} from "@ant-design/icons";
import {
  useEffect,
  useImperativeHandle,
  useState,
  Ref,
  useCallback,
  createRef,
  useRef,
} from "react";
import { uploadURL, downloadURL, wsURL, fileSizeLimit } from "./config";
import copy from "copy-to-clipboard";
import { v4 as uuidv4 } from "uuid";

function downloadContent(content: ArrayBuffer, filename: string) {
  const link = document.createElement("a");
  link.download = filename;
  link.style.display = "none";
  const blob = new Blob([content]);
  link.href = URL.createObjectURL(blob);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

function DownloadProgress({ cRef }: { cRef: Ref<any> }) {
  const [percent, setPercent] = useState(0);
  useImperativeHandle(cRef, () => ({
    setPercent: (val: number) => {
      setPercent(val);
    },
  }));
  return <div>{percent}%</div>;
}

function UploadProgress({ cRef }: { cRef: Ref<any> }) {
  const [percent, setPercent] = useState(0);
  useImperativeHandle(cRef, () => ({
    setPercent: (val: number) => {
      setPercent(val);
    },
  }));
  return <Progress percent={percent} />;
}

function App() {
  const [users, setUsers] = useState<User[]>([]);
  const [me, setMe] = useState<User>({
    id: "",
    name: "",
    device: "",
    is_mobile: false,
  });
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [ws, setWS] = useState<WebSocket>();
  const [roomIDVal, setRoomIDVal] = useState("");
  const [roomID, setRoomID] = useState(
    () => localStorage.getItem("roomID") || ""
  );
  const [msgModalOpen, setMsgModalOpen] = useState(false);
  const [msg, setMsg] = useState("");
  const [toUserID, setToUserID] = useState("");
  const [api, contextHolder] = notification.useNotification();
  const [messageApi, messageContextHolder] = message.useMessage();
  const sendFileStartRefs = useRef<{ [key: string]: boolean }>({});

  const copyMsg = useCallback(
    (key: string, text: string) => {
      copy(text);
      messageApi.success("复制成功");
      api.destroy(key);
    },
    [api, messageApi]
  );

  const handleReceiveMsg = useCallback(
    (data: ReceiveMsg) => {
      const key = uuidv4();
      const fromUser = users.find((item) => item.id === data.from);
      api.open({
        icon: <MessageOutlined />,
        message: fromUser?.name,
        description: data.msg,
        duration: 0,
        key,
        btn: <CopyOutlined onClick={() => copyMsg(key, data.msg)} />,
      });
    },
    [api, copyMsg, users]
  );

  const handleReceiveFileConfirm = useCallback(
    (data: ReceiveFileConfirm) => {
      const fromUser = users.find((item) => item.id === data.from);
      const key = uuidv4();
      let needRefuse = true;
      const useCacheDown = data.file_size <= fileSizeLimit;
      const downloadProgessRef = createRef<any>();
      api.open({
        key,
        icon: <FileOutlined />,
        message: fromUser?.name,
        description: data.filename,
        duration: null,
        onClose: () => {
          if (!needRefuse) {
            return;
          }
          const req: Req<TransferFileCancel> = {
            method: methodReceiveFileREFUSED,
            data: { download_id: data.download_id },
          };
          ws?.send(JSON.stringify(req));
        },
        btn: (
          <Flex gap={"middle"}>
            {useCacheDown ? <DownloadProgress cRef={downloadProgessRef} /> : ""}
            <DownloadOutlined
              onClick={() => {
                const downURL = `${downloadURL}?id=${data.download_id}`;
                if (useCacheDown && downloadProgessRef.current) {
                  const xhr = new XMLHttpRequest();
                  xhr.responseType = "arraybuffer";
                  xhr.onprogress = (e: ProgressEvent<EventTarget>) => {
                    if (e.lengthComputable) {
                      const p = Math.round((e.loaded * 100) / e.total);
                      console.log(p);
                      downloadProgessRef.current.setPercent(p);
                    }
                  };
                  xhr.onreadystatechange = () => {
                    if (xhr.readyState == 4) {
                      if (xhr.status !== 200) {
                        let reason = "";
                        if (xhr.response) {
                          reason = new TextDecoder("utf-8")
                            .decode(xhr.response)
                            .trim();
                        }
                        console.log(reason);

                        let errMsg = "下载失败";
                        if (reason) {
                          errMsg += `:${reason}`;
                        }
                        messageApi.error(errMsg);
                      } else {
                        console.log(xhr.responseType);
                        downloadContent(xhr.response, data.filename);
                        messageApi.success("下载成功");
                      }
                      needRefuse = false;
                      api.destroy(key);
                    }
                  };
                  xhr.open("GET", downURL);
                  xhr.send();
                } else {
                  open(downURL);
                  needRefuse = false;
                  api.destroy(key);
                }
              }}
            />
          </Flex>
        ),
      });
    },
    [api, messageApi, users, ws]
  );

  useEffect(() => {
    if (!roomID) {
      setIsModalOpen(true);
      return;
    }
    const websocket = new WebSocket(`${wsURL}?room_id=${roomID}`);
    websocket.onopen = () => {
      console.log("conn‌ect to server success");
    };
    websocket.onerror = function (error) {
      console.error("WebSocket 连接出现错误:", error);
    };
    websocket.onclose = function () {
      console.log("WebSocket 连接已经关闭。‌");
    };
    setWS(websocket);
    return () => {
      websocket.close();
    };
  }, [roomID]);

  useEffect(() => {
    if (!ws) return;
    ws.onmessage = function (event) {
      const resp = JSON.parse(event.data) as Resp<any>;
      console.dir(resp);
      switch (resp.method) {
        case methodFreshUsers:
          handleFreshUsers(resp.data);
          break;
        case methodReceiveMsg:
          handleReceiveMsg(resp.data);
          break;
        case methodSendFileStart:
          handleSendFileStart(resp.data);
          break;
        case methodReceiveFileConfirm:
          handleReceiveFileConfirm(resp.data);
          break;
      }
    };
  }, [handleReceiveFileConfirm, handleReceiveMsg, ws]);

  function fileUpload(
    e: React.ChangeEvent<HTMLInputElement>,
    from: string,
    to: string,
    toName: string
  ) {
    if (!e.target.files) {
      return;
    }
    const file = e.target.files[0];
    if (!file) {
      return;
    }

    console.log(toName);
    const xhr = new XMLHttpRequest();
    const downloadID = uuidv4();
    const uploadProgessRef = createRef<any>();
    let complete = false;
    api.open({
      key: downloadID,
      icon: <UploadOutlined />,
      message: toName,
      duration: null,
      description: (
        <>
          {file.name}
          <UploadProgress cRef={uploadProgessRef} />
        </>
      ),
      onClose: () => {
        const needCacelReq = sendFileStartRefs.current[downloadID];
        const tmp = { ...sendFileStartRefs.current };
        Reflect.deleteProperty(tmp, downloadID);
        sendFileStartRefs.current = tmp;
        if (complete) {
          return;
        }
        if (needCacelReq) {
          console.log("abort xhr");
          xhr.abort();
        } else {
          const req: Req<TransferFileCancel> = {
            method: methodSendFileCancel,
            data: { download_id: downloadID },
          };
          console.log(req);
          ws?.send(JSON.stringify(req));
        }
      },
    });

    xhr.upload.onprogress = (e: ProgressEvent<EventTarget>) => {
      if (e.lengthComputable && sendFileStartRefs.current[downloadID]) {
        const p = Math.round((e.loaded * 100) / e.total);
        uploadProgessRef.current.setPercent(p >= 0 ? p : 0);
      }
    };

    xhr.onreadystatechange = () => {
      if (xhr.readyState == 4) {
        if (xhr.status !== 200) {
          console.log(xhr.responseText);
          let errMsg = "传输失败";
          if (xhr.responseText.trim()) {
            errMsg += `:${xhr.responseText.trim()}`;
          }
          messageApi.error(errMsg);
        } else {
          messageApi.success("传输成功");
        }
        complete = true;
        api.destroy(downloadID);
      }
    };

    const query = new URLSearchParams({
      from,
      to,
      filename: file.name,
      room_id: roomID,
      download_id: downloadID,
    });

    xhr.open("POST", `${uploadURL}?${query.toString()}`);
    xhr.send(file);
  }

  function handleRoomIDInputChange(
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) {
    const val = e.target.value;
    setRoomIDVal(val);
    if (e.target.value.length === 6) {
      localStorage.setItem("roomID", val);
      ws?.close();
      setRoomID(val);
      setIsModalOpen(false);
      setRoomIDVal("");
    }
  }

  function handleJoinClick() {
    setIsModalOpen(true);
  }

  function handleCancel() {
    setIsModalOpen(false);
    setRoomIDVal("");
  }

  function handleMsgModalCancel() {
    setMsg("");
    setMsgModalOpen(false);
  }

  function handleMsgModalOK() {
    const req: Req<SendMsg> = {
      method: methodSendMsg,
      data: { to: toUserID, msg },
    };
    console.log(req);
    ws?.send(JSON.stringify(req));
    setMsg("");
    setMsgModalOpen(false);
  }

  function sendMsgClick(id: string) {
    setToUserID(id);
    setMsgModalOpen(true);
  }

  function handleFreshUsers(data: RegisterResp) {
    setUsers(data.users || []);
    setMe(data.me);
  }

  function handleSendFileStart(data: SendFileStartResp) {
    sendFileStartRefs.current = {
      ...sendFileStartRefs.current,
      [data.download_id]: true,
    };
  }

  return (
    <div id="container">
      {contextHolder}
      {messageContextHolder}
      <List
        header={
          <div id="header">
            <div>房间号: {roomID}</div>
            <div className="joinRoom" onClick={handleJoinClick}>
              加入其他房间
            </div>
          </div>
        }
        dataSource={users}
        itemLayout="horizontal"
        locale={{
          emptyText: <Empty description={"等待其他设备连接..."} />,
        }}
      >
        <List.Item actions={[<>我的设备</>]}>
          <List.Item.Meta
            avatar={
              <Avatar
                src={
                  me.is_mobile ? `/assets/mobile.svg` : `/assets/computer.svg`
                }
              />
            }
            title={me.name}
            description={me.device}
          />
        </List.Item>
        {users.map((item) => (
          <List.Item
            key={item.id}
            actions={[
              <div
                key="sendMsg"
                className="itemAction"
                onClick={() => {
                  sendMsgClick(item.id);
                }}
              >
                消息
              </div>,
              <div
                key="file"
                className="itemAction"
                style={{ cursor: "pointer" }}
              >
                <div>文件</div>
                <input
                  className="upload"
                  type="file"
                  name="upload"
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                    fileUpload(e, me.id, item.id, item.name)
                  }
                />
              </div>,
            ]}
          >
            <List.Item.Meta
              avatar={
                <Avatar
                  src={
                    item.is_mobile
                      ? `/assets/mobile.svg`
                      : `/assets/computer.svg`
                  }
                />
              }
              title={item.name}
              description={item.device}
            />
          </List.Item>
        ))}
      </List>
      <Modal
        title="请输入房间号"
        open={isModalOpen}
        onCancel={handleCancel}
        footer={[]}
        width={230}
      >
        <Input
          size="large"
          showCount
          maxLength={6}
          value={roomIDVal}
          onChange={handleRoomIDInputChange}
        />
      </Modal>
      <Modal
        title="请输入消息内容"
        open={msgModalOpen}
        onCancel={handleMsgModalCancel}
        onOk={handleMsgModalOK}
        cancelText="取消"
        okText="确定"
      >
        <Input.TextArea
          rows={4}
          value={msg}
          onChange={(val) => {
            setMsg(val.target.value);
          }}
        />
      </Modal>
    </div>
  );
}

export default App;
