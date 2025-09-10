import React from "react";
import styled from "styled-components";

const Panel = styled.div`
  background: #1a1a1a;
  border-radius: 8px;
  padding: 1.5rem;
  margin: 1rem 0;
  border: 1px solid #333;
`;

const Title = styled.div`
  color: #ffa500;
  font-size: 1rem;
  margin-bottom: 1rem;
  font-weight: 500;
`;

const AdvisorCard = styled.div`
  margin: 1rem 0;
  padding: 1rem;
  background: #252525;
  border-radius: 6px;
  border-left: 3px solid #4a9eff;
`;

const AdvisorHeader = styled.div`
  color: #ffffff;
  font-weight: 600;
  margin-bottom: 0.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
`;

const AdvisorNumber = styled.span`
  color: #4a9eff;
  font-weight: bold;
  margin-right: 0.5rem;
`;

const AdvisorAdvice = styled.div`
  color: #cccccc;
  line-height: 1.5;
  font-size: 0.95rem;
  padding-left: 1.5rem;
  position: relative;
  white-space: pre-wrap;

  &:before {
    content: "💭";
    position: absolute;
    left: 0;
    top: 0;
  }
`;

const AdvisorInfo = styled.div`
  flex: 1;
`;

const AdvisorName = styled.div`
  font-weight: bold;
  font-size: 0.9rem;
`;

const AdvisorTitle = styled.div`
  font-size: 0.8rem;
  opacity: 0.8;
`;

const AdvisorAvatar = styled.div`
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: ${(props) => getSpecialtyColor(props.specialty)};
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  margin-right: 0.75rem;
`;

const RecommendationBadge = styled.div`
  display: inline-block;
  background: rgba(255, 215, 0, 0.2);
  border: 1px solid #ffd700;
  color: #ffd700;
  padding: 0.3rem 0.6rem;
  border-radius: 15px;
  font-size: 0.8rem;
  font-weight: bold;
`;

function getSpecialtyColor(specialty) {
  const colors = {
    economy: "#4CAF50",
    security: "#F44336",
    diplomacy: "#2196F3",
    environment: "#8BC34A",
    domestic: "#FF9800",
    military: "#9C27B0",
    social: "#00BCD4",
    tech: "#607D8B",
  };
  return colors[specialty] || "#757575";
}

function getSpecialtyIcon(specialty) {
  const icons = {
    economy: "💰",
    security: "🛡️",
    diplomacy: "🤝",
    environment: "🌍",
    domestic: "🏛️",
    military: "⚔️",
    social: "👥",
    tech: "💻",
  };
  return icons[specialty] || "👤";
}

function getSpecialtyName(specialty) {
  const names = {
    economy: "Economy",
    security: "Security",
    diplomacy: "Diplomacy",
    environment: "Environment",
    domestic: "Domestic Policy",
    military: "Military Affairs",
    social: "Social Sphere",
    tech: "Technology",
  };
  return names[specialty] || specialty;
}

function AdvisorPanel({ advisors }) {
  if (!advisors || advisors.length === 0) return null;

  return (
    <Panel>
      <Title>💼 Your advisors weigh in:</Title>

      {advisors.map((advisor, index) => (
        <AdvisorCard key={index}>
          <AdvisorHeader>
            <AdvisorNumber>{index + 1}.</AdvisorNumber>
            👤 {advisor.advisorName || advisor.AdvisorName || advisor.name}
          </AdvisorHeader>
          <AdvisorAdvice>{advisor.advice || advisor.Advice}</AdvisorAdvice>
        </AdvisorCard>
      ))}
    </Panel>
  );
}

export default AdvisorPanel;
