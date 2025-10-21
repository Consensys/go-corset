(defcolumns (X :i16) (Y :i16))
(defconstraint c1 ()
  (begin
   (== 0 X)
   (== 0 Y)))
