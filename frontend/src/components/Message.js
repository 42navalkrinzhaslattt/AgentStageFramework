import React from "react";
import styled from "styled-components";
import ReactMarkdown from "react-markdown";

const MessageWrapper = styled.div`
  display: flex;
  align-items: flex-end;
  gap: 12px;
  margin-bottom: 16px;
  flex-direction: ${(props) => (props.isPlayer ? "row-reverse" : "row")};
`;

const MessageContainer = styled.div`
  flex: 1;
  min-width: 0;
  padding: 12px;
  background: ${(props) =>
    props.isPlayer ? "rgba(116, 185, 255, 0.15)" : "rgba(255, 255, 255, 0.05)"};
  border-radius: 12px;
  border: 1px solid
    ${(props) =>
      props.isPlayer ? "rgba(116, 185, 255, 0.3)" : "rgba(255, 255, 255, 0.1)"};
  backdrop-filter: blur(10px);
`;

const Avatar = styled.div`
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: ${(props) => props.color};
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
  border: 2px solid rgba(255, 255, 255, 0.2);
`;

const MessageHeader = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
`;

const Username = styled.span`
  font-weight: 600;
  color: ${(props) => props.color};
  font-size: 14px;
`;

const Timestamp = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  text-align: right;
  margin-top: 4px;
`;

const MessageText = styled.div`
  color: rgba(255, 255, 255, 0.9);
  line-height: 1.4;
  font-size: 14px;
  word-wrap: break-word;

  p {
    margin: 0 0 8px 0;
    &:last-child {
      margin-bottom: 0;
    }
  }

  strong {
    font-weight: 600;
    color: rgba(255, 255, 255, 1);
  }

  em {
    font-style: italic;
    color: rgba(255, 255, 255, 0.8);
  }

  code {
    background: rgba(255, 255, 255, 0.1);
    padding: 2px 4px;
    border-radius: 4px;
    font-family: "Courier New", monospace;
    font-size: 13px;
  }

  pre {
    background: rgba(255, 255, 255, 0.05);
    padding: 8px;
    border-radius: 6px;
    overflow-x: auto;
    margin: 8px 0;

    code {
      background: none;
      padding: 0;
    }
  }

  ul,
  ol {
    margin: 8px 0;
    padding-left: 20px;
  }

  li {
    margin: 4px 0;
  }
`;

const getSystemAvatar = (isSystem) => {
  if (isSystem) {
    return {
      color: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
      icon: "⚙️",
      name: "System",
      nameColor: "#a8b5ff",
    };
  }
  return {
    color: "#a8b5ff",
    icon: "🇺🇸",
    name: "Tronald Dump",
    nameColor: "#ff9ff3",
  };
};

const formatTime = (timestamp) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
};

const Message = ({ message, isSystem = false, time }) => {
  const avatarData = getSystemAvatar(isSystem);
  const timestamp = time || Date.now();
  const isPlayer = !isSystem;

  return (
    <MessageWrapper isPlayer={isPlayer}>
      <Avatar color={avatarData.color}>{avatarData.icon}</Avatar>
      <MessageContainer isPlayer={isPlayer}>
        <MessageHeader>
          <Username color={avatarData.nameColor}>{avatarData.name}</Username>
        </MessageHeader>
        <MessageText>
          <ReactMarkdown>{message}</ReactMarkdown>
        </MessageText>
        <Timestamp>{formatTime(timestamp)}</Timestamp>
      </MessageContainer>
    </MessageWrapper>
  );
};

export default Message;
