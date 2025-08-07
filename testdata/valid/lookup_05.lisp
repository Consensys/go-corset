(defcolumns (A :i16) (B :i16) (X :i16) (Y :i16))
(deflookup test (X Y) (A (* B 3)))
