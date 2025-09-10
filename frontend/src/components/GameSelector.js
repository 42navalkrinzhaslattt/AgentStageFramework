import React, { useState, useEffect } from "react";
import styled from "styled-components";

const SidebarContainer = styled.div`
  width: ${(props) => (props.$collapsed ? "80px" : "320px")};
  height: 100vh;
  background: #17212b;
  border-right: 1px solid #2c3e50;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: width 0.3s ease;

  @media (max-width: 768px) {
    width: 80px;
  }
`;

const SectionTitle = styled.div`
  padding: 16px 12px 8px 12px;
  color: #ffffff;
  font-size: 16px;
  font-weight: 500;
  text-align: center;
  background: #17212b;
  display: ${(props) => (props.$collapsed ? "none" : "block")};

  @media (max-width: 768px) {
    display: none;
  }
`;

const GamesList = styled.div`
  flex: 1;
  overflow-y: auto;
`;

const GameItem = styled.div`
  display: flex;
  align-items: center;
  padding: ${(props) => (props.$collapsed ? "12px 16px" : "12px")};
  cursor: pointer;
  background: ${(props) => (props.$active ? "#2b5278" : "transparent")};
  border-radius: 8px;
  margin: 2px 4px;
  opacity: ${(props) => (props.$coming ? 0.5 : 1)};
  transition: all 0.2s ease;
  justify-content: ${(props) => (props.$collapsed ? "center" : "flex-start")};

  @media (max-width: 768px) {
    padding: 12px 16px;
    justify-content: center;
  }

  &:hover {
    background: ${(props) =>
      props.$coming ? "transparent" : props.$active ? "#2b5278" : "#1e2329"};
  }
`;

const GameAvatar = styled.div`
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: ${(props) => props.color || "#5288c1"};
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
`;

const GameInfo = styled.div`
  flex: 1;
  min-width: 0;
  margin-left: 16px;
  display: ${(props) => (props.$collapsed ? "none" : "block")};

  @media (max-width: 768px) {
    display: none;
  }
`;

const GameName = styled.div`
  color: #ffffff;
  font-size: 15px;
  font-weight: 500;
  margin-bottom: 2px;
  display: flex;
  align-items: center;
  gap: 6px;
`;

const GameStatus = styled.div`
  color: ${(props) => (props.$coming ? "#8596a8" : "#7bb274")};
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 4px;
`;

const ComingSoonBadge = styled.span`
  background: #f39c12;
  color: #ffffff;
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 8px;
  font-weight: 500;
`;

const OnlineBadge = styled.span`
  width: 8px;
  height: 8px;
  background: #7bb274;
  border-radius: 50%;
`;

const games = [
  {
    id: "presidential-simulator",
    name: "Presidential Simulator",
    avatar: "🏛️",
    color: "#5288c1",
    status: "online",
    description: "Strategic political simulation game",
  },
  {
    id: "corporate-tycoon",
    name: "Corporate Tycoon",
    avatar: "🏢",
    color: "#e74c3c",
    status: "coming-soon",
    description: "Build your business empire",
  },
  {
    id: "space-commander",
    name: "Space Commander",
    avatar: "🚀",
    color: "#9b59b6",
    status: "coming-soon",
    description: "Explore the galaxy and command fleets",
  },
  {
    id: "medieval-kingdom",
    name: "Medieval Kingdom",
    avatar: "⚔️",
    color: "#f39c12",
    status: "coming-soon",
    description: "Rule your medieval realm",
  },
  {
    id: "cyber-detective",
    name: "Cyber Detective",
    avatar: "🕵️",
    color: "#1abc9c",
    status: "coming-soon",
    description: "Solve crimes in a cyberpunk world",
  },
];

function GameSelector({ selectedGame, onGameSelect }) {
  const [isCollapsed, setIsCollapsed] = useState(false);

  useEffect(() => {
    const checkScreenSize = () => {
      setIsCollapsed(window.innerWidth <= 768);
    };

    checkScreenSize();
    window.addEventListener("resize", checkScreenSize);

    return () => window.removeEventListener("resize", checkScreenSize);
  }, []);

  return (
    <SidebarContainer $collapsed={isCollapsed}>
      <SectionTitle $collapsed={isCollapsed}>Games</SectionTitle>
      <GamesList>
        {games.map((game) => (
          <GameItem
            key={game.id}
            $active={selectedGame === game.id}
            $coming={game.status === "coming-soon"}
            $collapsed={isCollapsed}
            onClick={() => game.status === "online" && onGameSelect(game.id)}
          >
            <GameAvatar color={game.color}>{game.avatar}</GameAvatar>
            <GameInfo $collapsed={isCollapsed}>
              <GameName>
                {game.name}
                {game.status === "coming-soon" && (
                  <ComingSoonBadge>Coming Soon</ComingSoonBadge>
                )}
              </GameName>
              <GameStatus $coming={game.status === "coming-soon"}>
                {game.status === "online" && <OnlineBadge />}
                {game.description}
              </GameStatus>
            </GameInfo>
          </GameItem>
        ))}
      </GamesList>
    </SidebarContainer>
  );
}

export default GameSelector;
