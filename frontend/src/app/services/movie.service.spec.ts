import { HttpClientTestingModule } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { HttpTestingController } from '@angular/common/http/testing';
import { environment } from '../../environments/environment';
import { MovieService } from './movie.service';

describe('MovieService', () => {
  let service: MovieService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule]
    });

    service = TestBed.inject(MovieService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('listMovies() sets only non-empty query params', () => {
    let response: any;

    service
      .listMovies({
        genre: 'Action',
        language: '',
        rating: undefined
      })
      .subscribe(res => {
        response = res;
      });

    const req = httpMock.expectOne(httpReq => {
      return (
        httpReq.url === `${environment.apiUrl}/movies` &&
        httpReq.params.get('genre') === 'Action' &&
        httpReq.params.get('language') === null &&
        httpReq.params.get('rating') === null
      );
    });

    req.flush({ movies: [] });

    expect(response.movies).toEqual([]);
  });
});

