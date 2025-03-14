export const methodFreshUsers = "FRESH_USERS";
export const methodSendMsg = "SEND_MSG";
export const methodReceiveMsg = "RECEIVE_MSG";
export const methodSendFileStart = "SEND_FILE_START";
export const methodSendFileCancel = "SEND_FILE_CANCEL";
export const methodReceiveFileConfirm = "RECEIVE_FILE_CONFIRM";
export const methodReceiveFileREFUSED = "RECEIVE_FILE_REFUSED";

export interface Req<T> {
  method: string;
  data: T;
}

export interface Resp<T> {
  code: string;
  msg: string;
  method: string;
  data: T;
}

export interface User {
  id: string;
  name: string;
  device: string;
  is_mobile: boolean;
}

export interface RegisterResp {
  me: User;
  users: User[];
}

export interface SendMsg {
  to: string;
  msg: string;
}

export interface ReceiveMsg {
  from: string;
  msg: string;
}

export interface ReceiveFileConfirm {
  from: string;
  filename: string;
  download_id: string;
  file_size: number;
}

export interface SendFileStartResp {
  download_id: string;
}

export interface TransferFileCancel {
  download_id: string;
}
