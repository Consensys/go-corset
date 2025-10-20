(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint c1 ()
  (if (== 0 X)
      (begin
       (== 0 Y)
       (== 0 Z))))
