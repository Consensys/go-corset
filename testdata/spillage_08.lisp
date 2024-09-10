(module m1)
(column S1)
(column A)
(vanish spills (* S1 (* A (~ (shift A 2)))))

(module m2)
(column S2)
(column B)
(vanish spills (* S2 (* B (~ (shift B -3)))))
