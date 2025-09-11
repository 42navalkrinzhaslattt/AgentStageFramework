import React from "react";
import styled from "styled-components";
import ReactMarkdown from "react-markdown";

const MessageWrapper = styled.div`
  display: flex;
  align-items: flex-start;
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
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
  text-transform: uppercase;
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

// Derive avatar/color based on message source (name/title) or system flag
function getAvatarData({ isSystem, name, title, titleColor }) {
  if (isSystem) {
    return {
      color: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
      icon: "⚙️",
      nameColor: titleColor || "#a8b5ff",
    };
  }

  const t = (title || "").toLowerCase();
  const n = (name || "").toLowerCase();

  // Event / News source
  if (n.includes("news desk") || t.includes("breaking") || t.includes("news")) {
    return {
      color: "linear-gradient(135deg, #2b5876 0%, #4e4376 100%)",
      icon: "🗞️",
      nameColor: titleColor || "#8b98a5",
    };
  }

  // Defense / Security roles
  if (t.includes("defense") || t.includes("security") || t.includes("homeland")) {
    return {
      color: "linear-gradient(135deg, #8a2b2b 0%, #b22222 100%)",
      icon: "🛡️",
      nameColor: titleColor || "#B22222",
    };
  }

  // National Security Advisor
  if (t.includes("national security advisor") || t.includes("nsa") || n.includes("advisor")) {
    return {
      color: "linear-gradient(135deg, #243949 0%, #517fa4 100%)",
      icon: "🧭",
      nameColor: titleColor || "#B22222",
    };
  }

  // Technology Advisor / Tech roles
  if (t.includes("technology") || t.includes("tech") || t.includes("cyber")) {
    return {
      color: "linear-gradient(135deg, #5b6fd6 0%, #7b68ee 100%)",
      icon: "💻",
      nameColor: titleColor || "#7B68EE",
    };
  }

  // Default person
  return {
    color: "#a8b5ff",
    icon: "👤",
    nameColor: titleColor || "#a8b5ff",
  };
}

function getInitials(name) {
  if (!name || typeof name !== 'string') return '?';
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean);
  if (parts.length >= 2) {
    const first = parts[0][0] || '';
    const last = parts[parts.length - 1][0] || '';
    return (first + last).toUpperCase();
  }
  const one = parts[0] || '';
  // Take first two characters for single-word names
  return one.slice(0, 2).toUpperCase();
}

const formatTime = (timestamp) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
};

// Render message content with special handling for the metrics summary line
function renderMessageBody(raw) {
  if (!raw || typeof raw !== 'string') return null;
  const lines = raw.split('\n');
  const firstLine = lines[0] || '';
  // Detect our metrics header line
  if (firstLine.startsWith('📈 Economy')) {
    const segments = firstLine.split(' | ');
    const rendered = segments.map((seg, idx) => {
      const s = seg.trim();
      // Capture label (with emoji) and value part (e.g., 28(+10) or 28)
      const m = s.match(/^(.*\D)\s(\d+(?:\([+-]?\d+\))?)$/);
      let label = s;
      let value = '';
      let color = '#FF6B6B';
      if (m) {
        label = m[1].trim();
        value = m[2].trim();
        const dm = value.match(/\(([+-]\d+)\)/);
        if (dm) {
          const d = parseInt(dm[1], 10);
          color = d > 0 ? '#3DDC84' : d < 0 ? '#FF6B6B' : '#A0A0A0';
        } else {
          // No brackets -> no change
          color = '#A0A0A0';
        }
      }
      return (
        <React.Fragment key={idx}>
          <span>{label} </span>
          <span style={{ color, fontWeight: 600 }}>{value}</span>
          {idx < segments.length - 1 ? ' | ' : ''}
        </React.Fragment>
      );
    });

    const rest = lines.slice(1).join('\n');
    return (
      <>
        <div>{rendered}</div>
        {rest && (
          <div style={{ marginTop: 8 }}>
            <ReactMarkdown>{rest}</ReactMarkdown>
          </div>
        )}
      </>
    );
  }

  // Default: render as markdown
  return <ReactMarkdown>{raw}</ReactMarkdown>;
}

const Message = ({ message, isSystem = false, isPlayer: isPlayerProp, time, name, title, titleColor, profilePicture }) => {
  const timestamp = time || Date.now();
  const isPlayer = typeof isPlayerProp === 'boolean' ? isPlayerProp : !isSystem;

  const avatar = getAvatarData({ isSystem, name, title, titleColor });

  const displayName = name || ""; // do not default to 'System' to allow anonymous messages
  const displayTitle = title;
  const displayTitleColor = titleColor || avatar.nameColor || "#8b98a5";
  const avatarText = getInitials(displayName);
  const avatarColor = avatar.color;
  const nameColor = avatar.nameColor || "#a8b5ff";
  const hasSender = !!displayName.trim();

  return (
    <MessageWrapper isPlayer={isPlayer}>
      {hasSender && <Avatar color={avatarColor}>{avatarText}</Avatar>}
      <MessageContainer isPlayer={isPlayer}>
        {hasSender && (
          <MessageHeader>
            <Username color={nameColor}>{displayName}</Username>
            {displayTitle && (
              <span style={{ color: displayTitleColor, fontSize: '12px', marginLeft: '8px' }}>
                {displayTitle}
              </span>
            )}
          </MessageHeader>
        )}
        <MessageText>
          {renderMessageBody(message)}
        </MessageText>
        <Timestamp>{formatTime(timestamp)}</Timestamp>
      </MessageContainer>
    </MessageWrapper>
  );
};

export default Message;
