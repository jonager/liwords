import React, { useMemo, useRef, useEffect } from 'react';
import { Card, Avatar, Row, Col } from 'antd';
import {
  GameTurn,
  GameEvent,
} from '../gen/macondo/api/proto/macondo/macondo_pb';
import { Board } from '../utils/cwgame/board';
import { millisToTimeStr } from '../store/timer_controller';
import { tilePlacementEventDisplay } from '../utils/cwgame/game_event';

type Props = {
  turns: Array<GameTurn>;
  currentTurn: GameTurn;
  board: Board;
};

type turnProps = {
  turn: GameTurn;
  board: Board;
};

type reducedPlayerInfo = {
  avatar?: string;
  nickname: string;
  fullName?: string;
};

type MoveEntityObj = {
  player: reducedPlayerInfo;
  coords: string;
  timeRemaining: string;
  rack: string;
  play: string;
  score: string;
  oldScore: number;
  cumulative: number;
};

const displaySummary = (evt: GameEvent, board: Board) => {
  // Handle just a subset of the possible moves here. These may be modified
  // later on.
  switch (evt.getType()) {
    case GameEvent.Type.EXCHANGE:
      return `Exch. ${evt.getExchanged()}`;

    case GameEvent.Type.PASS:
      return 'Passed.';

    case GameEvent.Type.TILE_PLACEMENT_MOVE:
      return tilePlacementEventDisplay(evt, board);

    case GameEvent.Type.UNSUCCESSFUL_CHALLENGE_TURN_LOSS:
      return 'Challenged!';
  }
  return '';
};

const Turn = (props: turnProps) => {
  const evts = props.turn.getEventsList();
  const memoizedTurn: MoveEntityObj = useMemo(() => {
    // Create a base turn, and modify it accordingly. This is memoized as we
    // don't want to do this relatively expensive computation all the time.
    console.log('computing memoized', evts);
    const turn = {
      player: {
        nickname: evts[0].getNickname(),
      },
      coords: evts[0].getPosition(),
      timeRemaining: millisToTimeStr(evts[0].getMillisRemaining(), false),
      rack: evts[0].getRack(),
      play: displaySummary(evts[0], props.board),
      score: `${evts[0].getScore()}`,
      cumulative: evts[0].getCumulative(),
      oldScore: evts[0].getCumulative() - evts[0].getScore(),
    };
    if (evts.length === 1) {
      return turn;
    }
    // Otherwise, we have to make some modifications.
    if (evts[1].getType() === GameEvent.Type.PHONY_TILES_RETURNED) {
      turn.score = '0';
      turn.cumulative = evts[1].getCumulative();
      turn.play = `(${turn.play})`;
    } else {
      // Otherwise, just add/subtract as needed.
      for (let i = 1; i < evts.length; i++) {
        switch (evts[i].getType()) {
          case GameEvent.Type.CHALLENGE_BONUS:
            turn.score = `${turn.score}+${evts[i].getBonus()}`;
            break;
          case GameEvent.Type.END_RACK_PENALTY:
            turn.score = `${turn.score}-${evts[i].getLostScore()}`;
            break;
          case GameEvent.Type.END_RACK_PTS:
            turn.score = `${turn.score}+${evts[i].getEndRackPoints()}`;
            break;
          case GameEvent.Type.TIME_PENALTY:
            turn.score = `${turn.score}-${evts[i].getLostScore()}`;
            break;
        }
        turn.cumulative = evts[i].getCumulative();
      }
    }
    return turn;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [evts]);

  return (
    <>
      <Row style={{ fontFamily: "'Source Code Pro', monospace" }}>
        <Col span={3}>
          <Avatar>{memoizedTurn.player.nickname[0].toUpperCase()}</Avatar>
        </Col>
        <Col span={4}>
          <strong>{memoizedTurn.coords}</strong> <br />
          {memoizedTurn.timeRemaining}
        </Col>
        <Col span={12}>
          <strong>{memoizedTurn.play}</strong> <br />
          {memoizedTurn.rack}
        </Col>
        <Col span={5}>
          {`${memoizedTurn.oldScore}+${memoizedTurn.score}`} <br />
          <strong>{memoizedTurn.cumulative}</strong>
        </Col>
      </Row>
    </>
  );
};

export const ScoreCard = (props: Props) => {
  const el = useRef<HTMLDivElement>(null);
  useEffect(() => {
    el.current?.scrollIntoView({ block: 'end', behavior: 'smooth' });
  }, [props.turns]);
  return (
    <Card
      style={{
        overflowY: 'scroll',
        maxHeight: 250,
      }}
      title={`Turn ${props.turns.length + 1}`}
      // eslint-disable-next-line jsx-a11y/anchor-is-valid
      extra={<a href="#">Notepad</a>}
    >
      {props.turns.map((t, idx) => (
        <Turn turn={t} board={props.board} key={`t_${idx + 0}`} />
      ))}
      {props.currentTurn.getEventsList().length ? (
        <Turn turn={props.currentTurn} board={props.board} />
      ) : null}
      <div id="dummy-end" ref={el} />
    </Card>
  );
};