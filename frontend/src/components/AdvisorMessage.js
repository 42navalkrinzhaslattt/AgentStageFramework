import React from "react";
import styled from "styled-components";

const AdvisorWrapper = styled.div`
  display: flex;
  align-items: flex-end;
  gap: 12px;
  margin-bottom: 16px;

  @media (max-width: 768px) {
    gap: 8px;
    margin-bottom: 12px;
  }
`;

const AdvisorMessageContainer = styled.div`
  flex: 1;
  min-width: 0;
  padding: 12px;
  padding-top: 20px; /* room for top-right title */
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  position: relative;

  @media (max-width: 768px) {
    padding: 8px;
    padding-top: 18px;
  }
`;

const AdvisorAvatar = styled.div`
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

const AdvisorHeader = styled.div`
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 4px;
`;

const MessageHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: flex-start; /* title moved to absolute top-right */
  margin-bottom: 4px;
`;

const NameSection = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
`;

const AdvisorName = styled.span`
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

const AdvisorTitle = styled.span`
  position: absolute;
  top: 8px;
  right: 10px;
  color: #8b98a5; /* fixed grey */
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  pointer-events: none;

  @media (max-width: 768px) {
    font-size: 11px;
    top: 6px;
    right: 8px;
  }
`;

const AdvisorAdvice = styled.div`
  color: #e1e5e9;
  font-size: 15px;
  line-height: 1.4;
  word-wrap: break-word;
  white-space: pre-wrap;

  @media (max-width: 768px) {
    font-size: 14px;
  }
`;

const MessageTime = styled.div`
  font-size: 12px;
  color: #72767d;
  margin-top: 2px;
`;

function getSpecialtyColor(specialty) {
  const colors = {
    economy: "#a8d8ea",
    security: "#ffb3ba",
    diplomacy: "#bae1ba",
    environment: "#b5e7e0",
    domestic: "#ffe4b5",
    military: "#d4b5d4",
    social: "#ffd1a3",
    tech: "#c7cedb",
  };
  return colors[specialty] || "#d1d5db";
}

function getSpecialtyNameColor(specialty) {
  const colors = {
    economy: "#4fc3f7",
    security: "#ff7043",
    diplomacy: "#66bb6a",
    environment: "#26a69a",
    domestic: "#ffb74d",
    military: "#ab47bc",
    social: "#ffa726",
    tech: "#78909c",
  };
  return colors[specialty] || "#90a4ae";
}

function getSpecialtyFromTitle(title) {
  const titleLower = title.toLowerCase();
  if (titleLower.includes('chief of staff') || titleLower.includes('chief staff')) {
    return 'domestic';
  }
  if (titleLower.includes('epa') || titleLower.includes('environment') || titleLower.includes('climate')) {
    return 'environment';
  }
  if (titleLower.includes('technology') || titleLower.includes('tech')) {
    return 'tech';
  }
  if (titleLower.includes('domestic') || titleLower.includes('policy')) {
    return 'domestic';
  }
  if (titleLower.includes('security') || titleLower.includes('national')) {
    return 'security';
  }
  if (titleLower.includes('economic') || titleLower.includes('finance')) {
    return 'economy';
  }
  if (titleLower.includes('diplomatic') || titleLower.includes('foreign')) {
    return 'diplomacy';
  }
  if (titleLower.includes('military') || titleLower.includes('defense')) {
    return 'military';
  }
  if (titleLower.includes('social') || titleLower.includes('welfare')) {
    return 'social';
  }
  return 'general';
}

function getSpecialtyIcon(specialty) {
  switch (specialty) {
    case "economy":
      return "💰";
    case "security":
      return "🛡️";
    case "diplomacy":
      return "🤝";
    case "environment":
      return "🌱";
    case "domestic":
      return "🏛️";
    case "military":
      return "⚔️";
    case "social":
      return "👥";
    case "technology":
      return "💻";
    default:
      return "🎯";
  }
}

function AdvisorMessage({ advisor, time }) {
  const advisorName =
    advisor.advisorName || advisor.AdvisorName || advisor.name || "Advisor";
  const advisorTitle = advisor.title || advisor.Title || "Presidential Advisor";
  const advice = advisor.advice || advisor.Advice || "";
  const specialty = advisor.specialty || getSpecialtyFromTitle(advisorTitle);
  const avatarColor = getSpecialtyColor(specialty);
  const nameColor = getSpecialtyNameColor(specialty);

  return (
    <AdvisorWrapper>
      <AdvisorAvatar color={avatarColor}>
        {getSpecialtyIcon(specialty)}
      </AdvisorAvatar>
      <AdvisorMessageContainer>
        <AdvisorTitle>{advisorTitle}</AdvisorTitle>
        <MessageHeader>
          <NameSection>
            <AdvisorName color={nameColor}>{advisorName}</AdvisorName>
          </NameSection>
        </MessageHeader>
        <AdvisorAdvice>{advice}</AdvisorAdvice>
        <Timestamp>{time}</Timestamp>
      </AdvisorMessageContainer>
    </AdvisorWrapper>
  );
}

export default AdvisorMessage;