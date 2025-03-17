(defcolumns (A :binary) (B :i16) (C :i16))

(defconstraint c1 ()
  (if (== A 1)
      (== 0 C)
      (== 0 B)))

(defconstraint c2 ()
  (if (== 1 A)
      (== 0 C)
      (== 0 B)))
