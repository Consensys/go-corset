(defcolumns (A :i16) (B :i16) (X :i16) (Y :i18))
(deflookup test (X Y) (A (* B 3)))
