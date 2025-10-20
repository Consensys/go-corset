(defcolumns (ST :i16) (A :i16) (B :i16))
(defconstraint spills ()
  (== 0
   (* ST A (~ (* (shift A 3) (shift B 2))))))
