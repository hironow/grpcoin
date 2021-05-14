package userdb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/grpcoin/grpcoin/api/grpcoin"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func makeTrade(p *Portfolio, action grpcoin.TradeAction, ticker string, quote, quantity *grpcoin.Amount) error {
	inCash := toFixed(p.CashUSD.V())
	pos, ok := p.Positions[ticker]
	if !ok {
		return status.Errorf(codes.InvalidArgument, "don't have a %s position in portfolio", ticker)
	}
	posN := toFixed(pos.V())
	// TODO support non-existing currencies by adding zero to positions map in the future

	cost := toFixed(quantity).Mul(toFixed(quote))

	posDelta := toFixed(quantity)

	if action == grpcoin.TradeAction_SELL {
		posDelta = posDelta.Neg()
		cost = cost.Neg()
	}

	finalCash := inCash.Sub(cost)
	finalPos := posN.Add(posDelta)
	if finalCash.IsNegative() {
		return status.Errorf(codes.InvalidArgument,
			"insufficient cash after transaction (%s)", finalCash)

	}
	if finalPos.IsNegative() {
		return status.Errorf(codes.InvalidArgument,
			"insufficient %s positions (%s) after transaction (current: %s)", ticker, finalPos, posN)
	}

	p.CashUSD = fromFixed(finalCash)
	p.Positions[ticker] = fromFixed(finalPos)
	return nil
}

func toFixed(a *grpcoin.Amount) decimal.Decimal {
	u, n := a.Units, a.Nanos
	if a.Units < 0 {
		u = -u
	}
	if a.Nanos < 0 {
		n = -n
	}
	s := fmt.Sprintf("%d.%09d", u, n)

	if a.Nanos < 0 && a.Units < 0 {
		s = "-" + s
	}
	return decimal.RequireFromString(s)
}

func fromFixed(i decimal.Decimal) Amount {
	s := i.Abs().StringFixed(9)
	p := strings.SplitN(s, ".", 2)
	ds := p[0]
	fs := p[1]
	d, _ := strconv.ParseInt(ds, 10, 64)
	f, _ := strconv.Atoi(fs)
	if i.IsNegative() {
		d, f = -d, -f
	}
	return Amount{Units: d, Nanos: int32(f)}
}