package main

import "mcprotproxy/mcnet"

//TODO read from file lol
const FAVICON = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAFKklEQVR4nN2beagVVRzHP9prhfaNDAzKooiwsqKN9oUwXwVBUZalIiEFlRoVJCYtFJr+IxE9WixoV4u0DbIiS83W10K+1Chfm88W217hEr/H797mjTNzzp2zvDv3888d5px7zu/3nZnfzDnndwahHNGxkUjsBowBLgYOBcSGL4BngbnALzHM6Bzf1vc7OJbXwJHAg0A3MBs4GdgL2FOPZ2tZB3BULKNCC7AdcDmwBPgQGA/sVFB/R2Ac8AHwLjAa2D6kgaEEGArcCawFHgdOLNHG8cBjwLfA3cABAez0KoA8y2cDC4A1wK3A3h7alTZuBlYDzwPnal9e8CGABLXrgS+BV4ELAt1Z0mY78DKwErgB2N1Ho2VJBrVZwMHOLtozDLjPR9BsVIBGg1ponIOmrQA+gloe0zVmuFIqaBYJECqopZmlt/T5wCJgi2N7DQXNLAG2Ba6JENSSbAYWAiNVjBnAz45tpoPmRPVtq0pJ9tFn6f7IQS2JXLkpwP7A1cAKD22KqHOApcC+yYKkAG16FUZ46NAHvcAjwLHAccCjwD+O7R6tPrbVTiQFGAscM2DuFvMecJXeFfJ8f+3Q1gj1tY+kAKNDWe+R9cA9wEHAKH2+ywTNuq9JAQ5vJk8NSNB8EThPY9XMBofRdV/TMaCKrAIm6+Mht/b7Fj5kxoBoMyKB+Bt4WOPYSEMX9ccmKcDvVfLWwCpD+ebaQVIA11dMM7GzwZbfsgT4o4UE2MVQninAT+Hsic5+hg7X1Q6SAnzTGr73MdRQ/mPtICnA2nD2RMc0FO6uHQzOOtkCDDO4sDpLgJUtJMBwQ3n9NZkU4JNw9kRliC64FPFplgAbHEdZzYJpOP8d8EOWAMJHLSDAqYbyfmOFtADL/NsTndMMHfabYUoLsLjavrOHxRpBv4ucFmCFxoKqcpFhAlem2d4qEmAT8GaFBbjQUP6GDpvrZKm1yL9dUZDb/xxDR6+kT2QJMC85Xq4QV+rSXRELbQSQUeHrFRRggqFcgl9X+mRewHjSj03ROAM4zNDZ3KyTeQI8U7EJklsM5f/mXdQ8ATbkKdaEyKrwWQazFuStNRa9M+dURIC7LOrMyCsoEuDzCgRDWf093VBnsS6tZWJa9p7qx84g7KArQiZyr75gEmBJ1sdDkzDdYuZHlsNfchFAmNaEzsuS+Y0W9SabFk9tBBAVn7K3LTi76ittG0NHz+kdXIht6sukJvkuGKTrfwca6smA5yabBm0F6NZnbqCZqkNeE7clZ359CIBmcy0fQAHGWMajZZp5bkUjAmzUzIo/y9nvRLtmpZr4SxOrNoUQAB1NTYri8v+062aKrVLcMpiomy+sKZP/94BmZMZA8hXnWzrfoZlkDVE2AXKCp/y9POQVd6/mK9rYKLmN15XpqKwAvTr/FmI9cYiOQaZY1u/SbNbeMp25pMB2ax5uj0MbaS4BPgZOsazfo5li6yzqZuKaA/yZJiS5TqVL3t8L+oVnWter0aMzQaZ8oEJ8JEEv1zvh15L/n6mRe1QD/6k531myzzq+ssCX6m37fYn/jrWM8jW+Ak7y4bzgMw2+06dhObwDnOAzl8H3PoA1auDTnttFvz/O9Bx0g2yEkE/lS4Fr08tQJZFR6GX6UVTqVVdEqJ0gW3RSdbjjkvtrmtj8hEfb+hF6K0yXbrAal0xNs0C+Ma7Qt0vQ9L0Ym6dlnfEhTWufpjn/eazXsfwhujvNdQOVkZi7xyUZ+3bN4ZPv9rfVYQlqciznpOwOHdaGB/gPfPP7b2bFbCkAAAAASUVORK5CYII="

func makeKickMsg(msg ProxyError) string {
	return string(`{"text":"` + msg + `", color: "red"}`)
}

func makeMotdMsg(msg ProxyError) string {
	return string(`{"description": {"text": "` + msg + `", "color": "red"},
			"favicon": "` + FAVICON + `",
			"version": {"name": "mcprot", "protocol": 9999},
			"players": {"max": 1, "online": 1}}`)
}

func WriteError(conn *mcnet.Conn, msg ProxyError, state State) {
	if state == StatusState {
		conn.WritePacket(0x00, mcnet.String(makeMotdMsg(msg)))
		conn.WritePacket(0x01, mcnet.VarShort(1))
		conn.Close()
	} else {
		conn.WritePacket(0x00, mcnet.String(makeKickMsg(msg)))
		conn.Close()
	}
}
