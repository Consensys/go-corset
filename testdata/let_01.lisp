(defcolumns (A :i16) (B :i16))
(defconstraint c1 ()
  (let ((C B))
    (if (== 0 A)
        (== 0 C))))
