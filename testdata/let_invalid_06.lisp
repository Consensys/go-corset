;;error:6:12-13:unknown symbol
(defpurefun (vanishes! x) (== 0 x))
(defcolumns (A :i16) (B :i16))

(defconstraint c1 ()
  (let ((B C))
    (if A
        (vanishes! B))))
