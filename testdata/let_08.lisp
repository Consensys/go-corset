(defcolumns (A :i16) (B :i16))
(defconstraint c1 ()
  (let ((B (+ B 1)))
    (if (== 0 A)
        (== B 1))))
