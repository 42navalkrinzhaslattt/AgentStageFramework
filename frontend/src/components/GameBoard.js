import React, { useState, useEffect, useRef, useCallback } from "react";
import styled from "styled-components";
import { useGame } from "../context/GameContext";
import MetricsPanel from "./MetricsPanel";
import EventCard from "./EventCard";
import AdvisorPanel from "./AdvisorPanel";
import ChoicePanel from "./ChoicePanel";
import LoadingSpinner from "./LoadingSpinner";
import MessageInput from "./MessageInput";
import AdvisorMessage from "./AdvisorMessage";
import MessageComponent from "./Message";

const TelegramContainer = styled.div`
  width: 100%;
  height: 100vh;
  background: #0f1419;
  display: flex;
  flex-direction: column;
  position: relative;
`;

const Header = styled.div`
  background: #17212b;
  display: flex;
  flex-direction: column;
  border-bottom: 1px solid #2c3e50;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
`;

const HeaderTop = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  margin-left: -16px;
  margin-right: -16px;
  padding-left: 16px;
  padding-right: 16px;

  @media (max-width: 768px) {
    padding: 8px 12px;
    margin-left: -12px;
    margin-right: -12px;
    padding-left: 12px;
    padding-right: 12px;
  }
`;

const PinnedMetricsContainer = styled.div`
  padding: 0 16px 12px 16px;

  @media (max-width: 768px) {
    padding: 0 12px 8px 12px;
  }
`;

const PinnedMetrics = styled.div`
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 8px 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #8b98a5;
  font-size: 13px;
  text-align: center;

  @media (max-width: 768px) {
    font-size: 11px;
    padding: 6px 8px;
  }
`;

const ChatInfo = styled.div`
  display: flex;
  align-items: center;
`;

const Avatar = styled.div`
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: ${(props) =>
    props.color || "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"};
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  margin-right: 12px;
  margin-left: 15px;
  font-size: 18px;
  border: 2px solid rgba(255, 255, 255, 0.2);

  @media (max-width: 768px) {
    width: 32px;
    height: 32px;
    font-size: 14px;
    margin-right: 8px;
    margin-left: 10px;
  }
`;

function getPresidentialSimulatorAvatar() {
  return {
    color: "linear-gradient(135deg, #5288c1 0%, #4a7ba7 100%)",
    icon: "🏛️",
  };
}

const TitleInfo = styled.div``;

const ChatTitle = styled.div`
  color: #ffffff;
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 2px;

  @media (max-width: 768px) {
    font-size: 14px;
  }
`;

const TurnInfo = styled.div`
  color: #8b98a5;
  font-size: 14px;

  @media (max-width: 768px) {
    font-size: 12px;
  }
`;

const MessagesArea = styled.div`
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 64px;

  @media (max-width: 768px) {
    padding: 12px 8px;
    gap: 8px;
    margin-bottom: 52px;
  }
`;

const Message = styled.div.withConfig({
  shouldForwardProp: (prop) => prop !== "isBot",
})`
  max-width: 85%;
  padding: 12px 16px;
  border-radius: 18px;
  background: ${(props) => (props.isBot ? "#2b5278" : "#0088cc")};
  color: white;
  align-self: ${(props) => (props.isBot ? "flex-start" : "flex-end")};
  font-size: 15px;
  line-height: 1.4;
  word-wrap: break-word;
  margin-bottom: 4px;
  white-space: pre-wrap;

  @media (max-width: 768px) {
    max-width: 95%;
    padding: 10px 12px;
    font-size: 14px;
    border-radius: 16px;
  }
`;

const MessageTime = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  margin-top: 4px;
  text-align: right;
`;

const SystemMessage = styled.div`
  text-align: center;
  color: #8b98a5;
  font-size: 13px;
  margin: 8px 0;
  padding: 8px;

  @media (max-width: 768px) {
    font-size: 12px;
    margin: 6px 0;
    padding: 6px;
  }
`;

const MetricsMessage = styled(Message)`
  background: #1a472a;
  max-width: 100%;
  align-self: center;
`;

const ErrorMessage = styled(Message)`
  background: #722f37;
  max-width: 100%;
  align-self: center;
`;

const GamePhase = {
  WAITING_FOR_EVENT: "waiting_for_event",
  SHOWING_EVENT: "showing_event",
  MAKING_CHOICE: "making_choice",
  EVALUATING: "evaluating",
  GAME_OVER: "game_over",
};

function GameBoard() {
  const {
    turn,
    maxTurns,
    currentTurn,
    metrics,
    loading,
    error,
    isComplete,
    newRound,
    evaluateChoice,
  } = useGame();

  const [gamePhase, setGamePhase] = useState(GamePhase.WAITING_FOR_EVENT);
  const [selectedChoice, setSelectedChoice] = useState(null);
  const [reasoning, setReasoning] = useState("");
  const [messages, setMessages] = useState([]);
  const [advisorMessages, setAdvisorMessages] = useState([]);
  const [isStartingRound, setIsStartingRound] = useState(false);
  const [inputMessage, setInputMessage] = useState("");
  const messagesEndRef = useRef(null);
  const hasStartedRound = useRef(false);
  const textareaRef = useRef(null);

  const addMessage = useCallback((text, isBot = false) => {
    const now = Date.now();
    const newMessage = {
      id: now + Math.random(),
      text,
      isBot,
      timestamp: now,
      time: new Date().toLocaleTimeString("ru-RU", {
        hour: "2-digit",
        minute: "2-digit",
      }),
    };
    setMessages((prev) => [...prev, newMessage]);
  }, []);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const handleSendMessage = async () => {
    if (inputMessage.trim()) {
      addMessage(inputMessage, false);
      const messageText = inputMessage;
      setInputMessage("");
      if (textareaRef.current) {
        textareaRef.current.style.height = "40px";
      }
      
      if (currentTurn) {
        try {
          await evaluateChoice(
            currentTurn.event.id,
            0,
            "user_message",
            messageText
          );
        } catch (err) {
          console.error("Failed to send message:", err);
          addMessage(`❌ Ошибка отправки сообщения: ${err.message}`, true);
        }
      }
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const handleInputChange = (e) => {
    setInputMessage(e.target.value);
    if (textareaRef.current) {
      textareaRef.current.style.height = "40px";
      textareaRef.current.style.height =
        textareaRef.current.scrollHeight + "px";
    }
  };

  const startNewRound = useCallback(async () => {
    if (isStartingRound || hasStartedRound.current) return;

    hasStartedRound.current = true;
    setIsStartingRound(true);
    setGamePhase(GamePhase.WAITING_FOR_EVENT);
    setAdvisorMessages([]);
    try {
      const result = await newRound();
      if (result.gameOver) {
        setGamePhase(GamePhase.GAME_OVER);
        addMessage("🏁 Game completed!", true);
      } else {
        setGamePhase(GamePhase.SHOWING_EVENT);
      }
    } catch (err) {
      console.error("Failed to start new round:", err);
      addMessage("❌ Error loading event", true);
    } finally {
      setIsStartingRound(false);
      hasStartedRound.current = false;
    }
  }, [isStartingRound, newRound, addMessage]);

  useEffect(() => {
    if (isComplete) {
      setGamePhase(GamePhase.GAME_OVER);
      addMessage("🏁 Game completed! Calculating results...", true);
      return;
    }

    if (!currentTurn && turn < maxTurns && !isStartingRound) {
      startNewRound();
    } else if (currentTurn) {
      setGamePhase(GamePhase.SHOWING_EVENT);
      addMessage(`📋 New event: ${currentTurn.event.title}`, true);
      addMessage(currentTurn.event.description, true);

      if (currentTurn.advisors && currentTurn.advisors.length > 0) {
        addMessage("💼 Your advisors weigh in:", true);
        const now = Date.now();
        const advisorMessagesWithTime = currentTurn.advisors.map(
          (advisor, index) => ({
            ...advisor,
            id: `advisor-${now}-${index}`,
            timestamp: now + index + 1,
            time: new Date().toLocaleTimeString("ru-RU", {
              hour: "2-digit",
              minute: "2-digit",
            }),
          })
        );
        setAdvisorMessages(advisorMessagesWithTime);
      }
    }
  }, [turn, maxTurns, currentTurn, isComplete, isStartingRound, startNewRound]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleChoiceSelect = (optionIndex, option) => {
    setSelectedChoice({ optionIndex, option });
    setGamePhase(GamePhase.MAKING_CHOICE);
  };

  const handleChoiceConfirm = async () => {
    if (!selectedChoice || !currentTurn) return;

    setGamePhase(GamePhase.EVALUATING);
    try {
      await evaluateChoice(
        currentTurn.event.id,
        selectedChoice.optionIndex,
        selectedChoice.option,
        reasoning
      );

      setSelectedChoice(null);
      setReasoning("");
      setGamePhase(GamePhase.WAITING_FOR_EVENT);
    } catch (err) {
      console.error("Failed to evaluate choice:", err);
      setGamePhase(GamePhase.MAKING_CHOICE);
    }
  };

  const handleChoiceCancel = () => {
    setSelectedChoice(null);
    setReasoning("");
    setGamePhase(GamePhase.SHOWING_EVENT);
  };

  const formatMetrics = (metrics) => {
    if (!metrics) return "No data";
    return `📊 Economy: ${metrics.economy}% | 🛡️ Security: ${metrics.security}% | 🤝 Diplomacy: ${metrics.diplomacy}% | 🌱 Environment: ${metrics.environment}%`;
  };

  return (
    <TelegramContainer>
      <Header>
        <HeaderTop>
          <ChatInfo>
            <Avatar color={getPresidentialSimulatorAvatar().color}>
              {getPresidentialSimulatorAvatar().icon}
            </Avatar>
            <TitleInfo>
              <ChatTitle>Presidential Simulator</ChatTitle>
              <TurnInfo>
                Turn {turn} of {maxTurns}
              </TurnInfo>
            </TitleInfo>
          </ChatInfo>
        </HeaderTop>
        {metrics && (
          <PinnedMetricsContainer>
            <PinnedMetrics>{formatMetrics(metrics)}</PinnedMetrics>
          </PinnedMetricsContainer>
        )}
      </Header>

      <MessagesArea>
        {[...messages.map(msg => ({...msg, type: 'message'})), ...advisorMessages.map(adv => ({...adv, type: 'advisor'}))]
          .sort((a, b) => {
            const timeA = a.timestamp || 0;
            const timeB = b.timestamp || 0;
            return timeA - timeB;
          })
          .map((item) => {
            if (item.type === 'message') {
              return (
                <MessageComponent
                  key={item.id}
                  message={item.text}
                  isSystem={item.isBot}
                  time={item.timestamp || Date.now()}
                />
              );
            } else {
              return (
                <AdvisorMessage
                  key={item.id}
                  advisor={item}
                  time={item.time}
                />
              );
            }
          })}

        {error && (
          <MessageComponent
            message={`❌ Error: ${error}`}
            isSystem={true}
            time={Date.now()}
          />
        )}

        {gamePhase === GamePhase.WAITING_FOR_EVENT && (
          <SystemMessage>⏳ Preparing new event...</SystemMessage>
        )}

        {gamePhase === GamePhase.EVALUATING && (
          <SystemMessage>
            🔄 Evaluating consequences of your decision...
          </SystemMessage>
        )}

        {gamePhase === GamePhase.GAME_OVER && (
          <SystemMessage>🏁 Game completed! Moving to results...</SystemMessage>
        )}

        <div ref={messagesEndRef} />
      </MessagesArea>

      <MessageInput
        value={inputMessage}
        onChange={handleInputChange}
        onKeyPress={handleKeyPress}
        onSend={handleSendMessage}
        textareaRef={textareaRef}
      />
    </TelegramContainer>
  );
}

export default GameBoard;
