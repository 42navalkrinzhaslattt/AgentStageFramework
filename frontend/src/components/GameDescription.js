import React, { useState } from "react";
import styled from "styled-components";

const DescriptionContainer = styled.div`
  flex: 1;
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 0 90px 0;
  background: #0f1419;
  color: #ffffff;
  text-align: center;
  box-sizing: border-box;
  overflow-y: auto;
  position: relative;

  @media (max-width: 768px) {
    padding: 20px 16px 100px 16px;
    height: 100vh;
    width: 100%;
  }
`;

const ContentWrapper = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  flex: 1;
  justify-content: center;
  min-height: calc(100vh - 200px);
`;

const ButtonContainer = styled.div`
  position: fixed;
  bottom: 20px;
  left: 320px;
  right: 0;
  display: flex;
  justify-content: center;
  padding: 0 10px;
  z-index: 1000;

  @media (max-width: 768px) {
    left: 80px;
    padding: 0 16px;
    bottom: 16px;
  }
`;

const GameIcon = styled.div`
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: ${(props) => props.color || "#5288c1"};
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  margin-bottom: 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);

  @media (max-width: 768px) {
    width: 80px;
    height: 80px;
    font-size: 32px;
    margin-bottom: 16px;
  }
`;

const GameTitle = styled.h1`
  font-size: 32px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #ffffff;

  @media (max-width: 768px) {
    font-size: 24px;
    margin: 0 0 12px 0;
  }
`;

const GameSubtitle = styled.p`
  font-size: 18px;
  color: #8596a8;
  margin: 0 0 32px 0;
  max-width: 600px;
  line-height: 1.5;

  @media (max-width: 768px) {
    font-size: 16px;
    margin: 0 0 24px 0;
    max-width: 100%;
  }
`;

const FeaturesList = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  max-width: 800px;
  margin-bottom: 40px;
  padding: 0 10px;

  @media (max-width: 900px) {
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  }

  @media (max-width: 860px) {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  @media (max-width: 768px) {
    grid-template-columns: 1fr;
    gap: 16px;
    margin-bottom: 24px;
    width: 100%;
  }

  @media (max-width: 600px) {
    grid-template-columns: 1fr;
  }
`;

const FeatureItem = styled.div`
  background: #17212b;
  padding: 20px;
  border-radius: 12px;
  border: 1px solid #2c3e50;
  text-align: left;

  @media (max-width: 768px) {
    padding: 16px;
    border-radius: 8px;
  }
`;

const FeatureIcon = styled.div`
  font-size: 24px;
  margin-bottom: 12px;
`;

const FeatureTitle = styled.h3`
  font-size: 16px;
  font-weight: 500;
  margin: 0 0 8px 0;
  color: #ffffff;
`;

const FeatureDescription = styled.p`
  font-size: 14px;
  color: #8596a8;
  margin: 0;
  line-height: 1.4;
`;

const ActionButton = styled.button`
  background: #5288c1;
  border: none;
  color: white;
  padding: 16px 32px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  transition: all 0.2s ease;

  &:hover {
    background: #4a7ba7;
    transform: translateY(-2px);
  }

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
  }

  @media (max-width: 768px) {
    padding: 14px 24px;
    font-size: 14px;
    width: 100%;
    max-width: 300px;
  }
`;

const ComingSoonButton = styled.button`
  background: #f39c12;
  border: none;
  color: white;
  padding: 16px 32px;
  border-radius: 8px;
  cursor: not-allowed;
  font-size: 16px;
  font-weight: 500;
  opacity: 0.8;
`;

const FullWidthButton = styled.button`
  width: 100%;
  max-width: 800px;
  background: #5288c1;
  border: none;
  color: white;
  padding: 20px;
  border-radius: 12px;
  cursor: pointer;
  font-size: 18px;
  font-weight: 500;
  transition: all 0.2s ease;

  &:hover {
    background: #4a7ba7;
    transform: translateY(-2px);
  }

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
  }

  @media (max-width: 768px) {
    padding: 16px;
    font-size: 16px;
  }
`;

const gameData = {
  "presidential-simulator": {
    title: "Presidential Simulator",
    subtitle:
      "Navigate complex political decisions and lead your nation through challenging times. Every choice shapes your legacy.",
    icon: "🏛️",
    color: "#5288c1",
    features: [
      {
        icon: "🎯",
        title: "Strategic Decisions",
        description:
          "Make critical choices that affect economy, security, and diplomacy",
      },
      {
        icon: "📊",
        title: "Real-time Metrics",
        description:
          "Monitor approval ratings, stability, and various national indicators",
      },
      {
        icon: "🤝",
        title: "Advisors",
        description:
          "Get expert advice from specialized AI advisors in different fields",
      },
      {
        icon: "📰",
        title: "Dynamic Events",
        description:
          "Respond to breaking news and unexpected crises as they unfold",
      },
    ],
    available: true,
  },
  "corporate-tycoon": {
    title: "Corporate Tycoon",
    subtitle:
      "Build your business empire from the ground up. Manage resources, compete with rivals, and dominate the market.",
    icon: "🏢",
    color: "#e74c3c",
    features: [
      {
        icon: "💼",
        title: "Business Management",
        description: "Oversee operations, finances, and strategic planning",
      },
      {
        icon: "📈",
        title: "Market Analysis",
        description: "Study trends and make data-driven investment decisions",
      },
      {
        icon: "🤖",
        title: "AI Competitors",
        description: "Face intelligent rival companies with unique strategies",
      },
      {
        icon: "🌍",
        title: "Global Expansion",
        description: "Expand your business across different markets worldwide",
      },
    ],
    available: false,
  },
};

function GameDescription({ gameId, onStartGame }) {
  const game = gameData[gameId] || gameData["presidential-simulator"];

  const handleStartClick = () => {
    if (game.available) {
      onStartGame();
    }
  };

  return (
    <DescriptionContainer>
      <ContentWrapper>
        <GameTitle>{game.title}</GameTitle>
        <GameSubtitle>{game.subtitle}</GameSubtitle>

        <FeaturesList>
          {game.features.map((feature, index) => (
            <FeatureItem key={index}>
              <FeatureIcon>{feature.icon}</FeatureIcon>
              <FeatureTitle>{feature.title}</FeatureTitle>
              <FeatureDescription>{feature.description}</FeatureDescription>
            </FeatureItem>
          ))}
        </FeaturesList>
      </ContentWrapper>

      <ButtonContainer>
        {game.available ? (
          <FullWidthButton onClick={handleStartClick}>Start</FullWidthButton>
        ) : (
          <ComingSoonButton>Coming Soon</ComingSoonButton>
        )}
      </ButtonContainer>
    </DescriptionContainer>
  );
}

export default GameDescription;
