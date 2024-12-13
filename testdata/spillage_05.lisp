(defpurefun ((vanishes! :@loob) x) x)

(module m1)
(defcolumns S1 A)
(defconstraint spills ()
  (vanishes! (* S1 (* A (~ (shift A 2))))))

(module m2)
(defcolumns S2 B)
(defconstraint spills ()
  (vanishes! (* S2 (* B (~ (shift B 3))))))
