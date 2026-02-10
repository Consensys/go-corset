(module f)
(defcolumns (X :i16) (Y :u1) (R :i16))
(deflookup l1 (X Y R) (id.x 1 id.r))
