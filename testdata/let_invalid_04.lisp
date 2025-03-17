;;error:6:15-18:malformed let assignment
(defpurefun (vanishes! x) (== 0 x))
(defcolumns (A :i16) (B :i16))

(defconstraint c1 ()
  (let ((C B) (D))
    (if A
        (vanishes! C))))
