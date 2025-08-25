(defpurefun (Id A) (if (== A 0) A A))
;;
(defcolumns (X :i16) (Y :i16))
(deflookup test (Y) ((Id X)))
