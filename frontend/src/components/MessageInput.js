import React from "react";
import styled from "styled-components";

const MessageInputContainer = styled.div`
  position: fixed;
  bottom: 0;
  left: 320px;
  right: 0;
  background: #17212b;
  border-top: 1px solid #2c3e50;
  padding: 12px 16px;
  display: flex;
  align-items: flex-end;
  gap: 12px;
  z-index: 1000;

  @media (max-width: 768px) {
    left: 80px;
    padding: 8px 12px;
    gap: 8px;
  }
`;

const TextArea = styled.textarea`
  flex: 1;
  background: transparent;
  border: none;
  padding: 10px 16px;
  color: white;
  font-size: 15px;
  outline: none;
  resize: none;
  min-height: 40px;
  max-height: 120px;
  overflow-y: auto;
  font-family: inherit;
  line-height: 1.4;

  &::-webkit-scrollbar {
    display: none;
  }

  -ms-overflow-style: none;
  scrollbar-width: none;

  &::placeholder {
    color: #8b98a5;
  }

  @media (max-width: 768px) {
    font-size: 14px;
    padding: 8px 12px;
  }
`;

const SendButton = styled.button`
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #5288c1;
  border: none;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  outline: none;
  padding: 8px;

  &:hover {
    transform: scale(1.1);
    background: #4a7bb7;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }

  svg {
    width: 20px;
    height: 20px;
    fill: white;
  }

  @media (max-width: 768px) {
    width: 36px;
    height: 36px;
  }
`;

const StartButton = styled.button`
  flex: 1;
  background: #5288c1;
  border: none;
  padding: 12px 24px;
  color: white;
  font-size: 16px;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  outline: none;

  &:hover {
    background: #4a7bb7;
    transform: translateY(-1px);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }

  @media (max-width: 768px) {
    font-size: 14px;
    padding: 10px 20px;
  }
`;

function MessageInput({
  value,
  onChange,
  onKeyPress,
  onSend,
  placeholder = "Напишите сообщение...",
  disabled = false,
  textareaRef,
}) {
  return (
    <MessageInputContainer>
      <TextArea
        ref={textareaRef}
        placeholder={placeholder}
        value={value}
        onChange={onChange}
        onKeyPress={onKeyPress}
        disabled={disabled}
      />
      <SendButton onClick={onSend} disabled={!value?.trim() || disabled}>
        <svg viewBox="0 0 24 24">
          <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
        </svg>
      </SendButton>
    </MessageInputContainer>
  );
}

export default MessageInput;
