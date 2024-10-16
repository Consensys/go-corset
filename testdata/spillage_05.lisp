(module m1)
(defcolumns S1 A)
(vanish spills (* S1 (* A (~ (shift A 2)))))

(module m2)
(defcolumns S2 B)
(vanish spills (* S2 (* B (~ (shift B 3)))))
