import React from "react";
import styled from "styled-components";
import { useGame } from "../context/GameContext";

const TelegramContainer = styled.div`
  width: 100vw;
  height: 100vh;
  background: #0f1419;
  display: flex;
  flex-direction: column;
`;

const Header = styled.div`
  background: #17212b;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  border-bottom: 1px solid #2c3e50;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
`;

const Avatar = styled.div`
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: ${props => props.color || 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'};
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  margin-right: 12px;
  font-size: 18px;
  border: 2px solid rgba(255, 255, 255, 0.2);
`;

function getPresidentialSimulatorAvatar() {
  return {
    color: 'linear-gradient(135deg, #5288c1 0%, #4a7ba7 100%)',
    icon: '🏛️'
  };
}

const ChatInfo = styled.div`
  flex: 1;
`;

const ChatTitle = styled.div`
  color: #ffffff;
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 2px;
`;

const ChatStatus = styled.div`
  color: #8b98a5;
  font-size: 14px;
`;

const MessagesArea = styled.div`
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
`;

const Message = styled.div.withConfig({
  shouldForwardProp: (prop) => prop !== "isBot",
})`
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 18px;
  background: ${(props) => (props.isBot ? "#2b5278" : "#5288c1")};
  color: white;
  align-self: ${(props) => (props.isBot ? "flex-start" : "flex-end")};
  position: relative;
  font-size: 15px;
  line-height: 1.4;
  word-wrap: break-word;
`;

const MessageTime = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  margin-top: 4px;
  text-align: right;
`;

const InputArea = styled.div`
  background: #17212b;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-top: 1px solid #2c3e50;
`;

const StartButton = styled.button`
  background: #5288c1;
  border: none;
  color: white;
  padding: 10px 20px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background 0.2s ease;

  &:hover {
    background: #4a7ba7;
  }

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

function StartScreen() {
  const { startGame, loading, error } = useGame();

  const handleStart = async () => {
    try {
      await startGame();
    } catch (err) {
      console.error("Failed to start game:", err);
    }
  };

  const getCurrentTime = () => {
    return new Date().toLocaleTimeString("ru-RU", {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <TelegramContainer>
      <Header>
        <Avatar color={getPresidentialSimulatorAvatar().color}>
          {getPresidentialSimulatorAvatar().icon}
        </Avatar>
        <ChatInfo>
          <ChatTitle>Presidential Simulator</ChatTitle>
          <ChatStatus>online</ChatStatus>
        </ChatInfo>
      </Header>

      <MessagesArea>
        <Message isBot={true}>
          Welcome to Presidential Simulator! 🏛️
          <MessageTime>{getCurrentTime()}</MessageTime>
        </Message>

        <Message isBot={true}>
          You are the president of a country facing multiple crises. Each turn
          you will need to make difficult decisions that will affect the
          economy, security, diplomacy and environment.
          <MessageTime>{getCurrentTime()}</MessageTime>
        </Message>

        <Message isBot={true}>
          Your advisors will give recommendations, but the final choice is
          yours. Ready to start governing the country?
          <MessageTime>{getCurrentTime()}</MessageTime>
        </Message>

        {error && (
          <Message isBot={true}>
            ❌ Error: {error}
            <MessageTime>{getCurrentTime()}</MessageTime>
          </Message>
        )}
      </MessagesArea>

      <InputArea>
        <StartButton onClick={handleStart} disabled={loading}>
          {loading ? "⏳ Initializing..." : "🚀 Start Game"}
        </StartButton>
      </InputArea>
    </TelegramContainer>
  );
}

export default StartScreen;
