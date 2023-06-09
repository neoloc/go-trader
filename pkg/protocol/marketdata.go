package protocol

import (
	"bytes"
	. "github.com/neoloc/go-trader/pkg/common"
)

// very simplified structure, only one book and associated trades per UDP packet, and it contains the complete book
// currently it all needs to fit in a single packet or it won't work, although straightforward to send additional trade packets

// MaxMsgSize is the maximum length of a multicast message
const MaxMsgSize = 1024

func EncodeMarketEvent(w *bytes.Buffer, book *Book, trades []Trade) {
	PutVarint(w, book.Instrument.ID())
	if book != nil {
		w.WriteByte(1) // has book
		encodeBook(w, book)
	} else {
		w.WriteByte(0) // no book
	}
	encodeTrades(w, trades)
}

func DecodeMarketEvent(r *bytes.Buffer) (*Book, []Trade) {
	instrumentId, _ := ReadVarint(r)
	instrument := IMap.GetByID(instrumentId)

	if instrument == nil {
		return nil, nil
	}

	hasBook, _ := r.ReadByte()
	var book *Book
	if hasBook == 1 {
		book = decodeBook(r, instrument)
	}
	trades := decodeTrades(r, instrument)
	return book, trades
}

func encodeBook(buf *bytes.Buffer, book *Book) {
	PutUvarint(buf, book.Sequence)

	encodeLevels(buf, book.Bids)
	encodeLevels(buf, book.Asks)
}

func decodeBook(r *bytes.Buffer, instrument Instrument) *Book {
	book := new(Book)

	sequence, _ := ReadUvarint(r)

	book.Instrument = instrument
	book.Sequence = sequence

	book.Bids = decodeLevels(r)
	book.Asks = decodeLevels(r)

	return book
}

func encodeLevels(w *bytes.Buffer, levels []BookLevel) {
	w.WriteByte(byte(len(levels)))
	for _, level := range levels {
		EncodeDecimal(w, level.Price)
		EncodeDecimal(w, level.Quantity)
	}
}

func decodeLevels(r ByteReader) []BookLevel {
	n, _ := r.ReadByte()
	levels := make([]BookLevel, n)
	for i := 0; i < int(n); i++ {
		price := DecodeDecimal(r)
		qty := DecodeDecimal(r)
		levels[i] = BookLevel{Price: price, Quantity: qty}
	}
	return levels
}

// this will blow up if any given match generates a ton of trades...
func encodeTrades(buf *bytes.Buffer, trades []Trade) {
	buf.WriteByte(byte(len(trades)))
	for _, v := range trades {
		EncodeDecimal(buf, v.Quantity)
		EncodeDecimal(buf, v.Price)
		EncodeString(buf, v.ExchangeID)
		EncodeTime(buf, v.TradeTime)
	}
}

func decodeTrades(r *bytes.Buffer, instrument Instrument) []Trade {
	n, _ := r.ReadByte() // read length
	trades := make([]Trade, n)
	for i := 0; i < int(n); i++ {
		qty := DecodeDecimal(r)
		price := DecodeDecimal(r)
		exchangeID := DecodeString(r)
		tradeTime := DecodeTime(r)

		trades[i] = Trade{Instrument: instrument, Price: price, Quantity: qty, ExchangeID: exchangeID, TradeTime: tradeTime}
	}
	return trades
}

type ReplayRequest struct {
	// Start is inclusive, and End is exclusive
	Start, End uint64
}
